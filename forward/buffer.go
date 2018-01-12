// Copyright 2018 pixiv Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package forward

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// A Buffer of message stream.
//
// The implementation is strongly influenced by Fluentd.
// The buffer structure is a queue of chunks like following:
//
//           stage:
//           map[tag]chunk             queue:
//           +---------------+         []chunk
//           | tag A:        |         +-----------+
//           | +-----------+ | enqueue | +-------+ | takeback (if writing failed)
//           | |   chunk   |------------>| chunk |<----------------------------+
//    append | |+---------+| |         | +-------+ |                           |
//   ---------->|  entry  || |         | +-------+ |       dequeued:           |
//           | ||    :    || |         | | chunk | |       map[unique_id]chunk |
//           | |+---------+| |         | +-------+ |       +-----------+       |
//           | +-----------+ |         |     :     |       | id XXX:   |       |
//           |               |         | +-------+ | write | +-------+ |       |
//           | tag B:        |         | | chunk |-----+---->| chunk |---------+
//           | +-----------+ |         | +-------+ |   |   | +-------+ |
//           | |   chunk   | |         +-----------+   |   | id YYY:   | purge
//           | +-----------+ |                         |   | +-------+ | (if writing suceeded)
//           |       :       |                         |   | | chunk |--------->
//           |               |                         |   | +-------+ |
//           +---------------+                         |   +-----------+
//                                                     |
//                                                     +-------> output
//
type Buffer interface {
	Append(string, Entry) error
	Enqueue() error
	Dequeue(context.Context) <-chan *Chunk
	Takeback(*Chunk) error
	Purge(*Chunk) error
	Stat() BufferStat
}

type MemoryBuffer struct {
	config  *Config
	stage   map[string]*Chunk
	queue   chan *Chunk
	stageMu sync.Mutex
	pool    *chunkPool
	closed  bool
	stat    BufferStat
}

var _ Buffer = &MemoryBuffer{}

// NewMemoryBuffer creates a memory buffer that writes chunks into the connection rw.
func NewMemoryBuffer(config *Config) *MemoryBuffer {
	buffer := &MemoryBuffer{
		config: config,
		stage:  make(map[string]*Chunk),
		queue:  make(chan *Chunk, config.BufferQueueLimit),
		pool:   globalChunkPool,
	}
	return buffer
}

// Append appends an entry of a tag into the buffer.
func (b *MemoryBuffer) Append(tag string, entry Entry) error {
	b.stageMu.Lock()
	defer b.stageMu.Unlock()

	// get the chunk
	chunk, exist := b.stage[tag]
	if !exist {
		chunk = b.newStageChunk(tag)
	}

	// enqueue the chunk if the entries size will exceed BufferChunkLimit
	if chunk.Msgsize()+entry.Msgsize() > b.config.BufferChunkLimit && b.config.BufferChunkLimit > 0 {
		err := b.enqueue(tag)
		if err != nil {
			return err
		}
		chunk = b.newStageChunk(tag)
	}
	return b.appendEntry(chunk, entry)
}

// NOTE: call in stage lock
func (b *MemoryBuffer) newStageChunk(tag string) *Chunk {
	chunk := b.pool.Get()
	chunk.Tag = tag
	b.stage[tag] = chunk
	b.stat.incrStage()
	return chunk
}

func (b *MemoryBuffer) appendEntry(chunk *Chunk, entry Entry) error {
	var err error
	chunk.Entries, err = entry.MarshalMsg(chunk.Entries)
	return errors.Wrap(err, "failed to append an entry into a chunk")
}

// Enqueue pushes all staged chunks into the write queue.
func (b *MemoryBuffer) Enqueue() error {
	b.stageMu.Lock()
	defer b.stageMu.Unlock()

	for tag := range b.stage {
		err := b.enqueue(tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// NOTE: call in stage lock
func (b *MemoryBuffer) enqueue(tag string) error {
	b.queue <- b.stage[tag]
	delete(b.stage, tag)
	b.stat.incrQueue()
	b.stat.decrStage()
	return nil
}

func (b *MemoryBuffer) Dequeue(ctx context.Context) <-chan *Chunk {
	ch := make(chan *Chunk)
	go func(queue <-chan *Chunk, out <-chan *Chunk) {
		for {
			select {
			case <-ctx.Done():
				return
			case chunk := <-b.queue:
				ch <- chunk
				b.stat.decrQueue()
				b.stat.incrDequeued()
			}
		}
	}(b.queue, ch)
	return ch
}

func (b *MemoryBuffer) Purge(chunk *Chunk) error {
	b.stat.decrDequeued()
	b.pool.Put(chunk)
	return nil
}

func (b *MemoryBuffer) Takeback(chunk *Chunk) error {
	b.queue <- chunk
	b.stat.incrQueue()
	return nil
}

// Stat returns buffer statistics instant.
func (b *MemoryBuffer) Stat() BufferStat {
	return b.stat.clone()
}

type chunkPool struct {
	pool sync.Pool
}

var globalChunkPool *chunkPool

func init() {
	globalChunkPool = newChunkPool()
}

func newChunkPool() *chunkPool {
	p := &chunkPool{
		pool: sync.Pool{
			New: newChunk,
		},
	}
	return p
}

func newChunk() interface{} {
	chunk := &Chunk{
		Entries: nil,
		Option:  make(map[string]interface{}),
	}
	return chunk
}

func (p *chunkPool) Get() *Chunk {
	return p.pool.Get().(*Chunk)
}

func (p *chunkPool) Put(c *Chunk) {
	c.Tag = c.Tag[:0]
	c.Entries = c.Entries[:0]
	c.Option = make(map[string]interface{})
	p.pool.Put(c)
}

// BufferStat holds statistics of a buffer.
type BufferStat struct {
	Stage    int64
	Queue    int64
	Dequeued int64
}

func (s *BufferStat) clone() BufferStat {
	return BufferStat{
		Stage:    atomic.LoadInt64(&s.Stage),
		Queue:    atomic.LoadInt64(&s.Queue),
		Dequeued: atomic.LoadInt64(&s.Dequeued),
	}
}

func (s *BufferStat) incrStage() {
	atomic.AddInt64(&s.Stage, 1)
}

func (s *BufferStat) decrStage() {
	atomic.AddInt64(&s.Stage, -1)
}

func (s *BufferStat) incrQueue() {
	atomic.AddInt64(&s.Queue, 1)
}

func (s *BufferStat) decrQueue() {
	atomic.AddInt64(&s.Queue, -1)
}

func (s *BufferStat) incrDequeued() {
	atomic.AddInt64(&s.Dequeued, 1)
}

func (s *BufferStat) decrDequeued() {
	atomic.AddInt64(&s.Dequeued, -1)
}
