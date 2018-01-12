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
	"io"
	"sync"
	"sync/atomic"
	"time"

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
type Buffer struct {
	rw       io.ReadWriteCloser
	config   *Config
	stage    map[string]*Chunk
	queue    chan *Chunk
	dequeued map[ChunkID]*Chunk
	stageMu  sync.Mutex
	pool     *chunkPool
	closed   bool
	shutdown bool
	stat     BufferStat
}

// NewBuffer creates a new buffer that writes chunks into the connection rw.
func NewBuffer(ctx context.Context, output io.ReadWriteCloser, config *Config) *Buffer {
	buffer := &Buffer{
		rw:       output,
		config:   config,
		stage:    make(map[string]*Chunk),
		queue:    make(chan *Chunk, config.BufferQueueLimit),
		dequeued: make(map[ChunkID]*Chunk),
		pool:     globalChunkPool,
	}
	buffer.start(ctx)
	return buffer
}

// Append appends an entry of a tag into the buffer.
func (b *Buffer) Append(tag string, entry Entry) error {
	b.stageMu.Lock()
	defer b.stageMu.Unlock()
	if b.shutdown {
		return errors.New("buffer is shutting down")
	}

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

	var err error
	chunk.Entries, err = entry.MarshalMsg(chunk.Entries)
	err = errors.Wrap(err, "marshal entry")
	return err
}

// NOTE: call in stage lock
func (b *Buffer) newStageChunk(tag string) *Chunk {
	chunk := b.pool.Get()
	chunk.Tag = tag
	b.stage[tag] = chunk
	b.stat.incrStage()
	return chunk
}

// Enqueue pushes the staged chunk into the write queue.
func (b *Buffer) Enqueue(tag string) error {
	b.stageMu.Lock()
	defer b.stageMu.Unlock()

	if _, exist := b.stage[tag]; !exist {
		return nil
	}
	err := b.enqueue(tag)
	if err != nil {
		return err
	}
	return nil
}

// NOTE: call in stage lock
func (b *Buffer) enqueue(tag string) error {
	b.queue <- b.stage[tag]
	delete(b.stage, tag)
	b.stat.incrQueue()
	b.stat.decrStage()
	return nil
}

func (b *Buffer) start(ctx context.Context) {
	go b.writeLoop(ctx)
	go b.flushLoop(ctx)
}

func (b *Buffer) writeLoop(ctx context.Context) {
	buf := make([]byte, b.config.BufferChunkLimit)
	retryInterval := b.config.RetryInterval
	for {
		var chunk *Chunk
		select {
		case <-ctx.Done():
			return
		case chunk = <-b.queue:
		}
		b.stat.decrQueue()
		err := b.write(chunk, buf)
		if err != nil {
			b.queue <- chunk

			// wait for retry
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
			}
			retryInterval = retryInterval * 2
			if retryInterval > b.config.MaxRetryInterval {
				retryInterval = b.config.MaxRetryInterval
			}
			continue
		}
		retryInterval = b.config.RetryInterval
	}
}

// NOTE: call in queue block
func (b *Buffer) write(chunk *Chunk, buf []byte) error {
	var err error
	buf, err = chunk.MarshalMsg(buf[:0])
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}
	_, err = b.rw.Write(buf)
	if err != nil {
		return err
	}
	b.pool.Put(chunk)
	return nil
}

func (b *Buffer) flushLoop(ctx context.Context) {
	t := time.NewTicker(b.config.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		b.stageMu.Lock()
		for tag := range b.stage {
			err := b.enqueue(tag)
			if err != nil {
				continue
			}
		}
		b.stageMu.Unlock()
	}
}

// Stat returns buffer statistics instant.
func (b *Buffer) Stat() BufferStat {
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
