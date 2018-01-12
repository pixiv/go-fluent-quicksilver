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

package forward_test

import (
	"context"
	"testing"
	"time"

	"github.com/pixiv/go-fluent-quicksilver/forward"
)

var tag = "test"
var entry = forward.Entry{
	Time:   forward.NewEventTime(time.Now()),
	Record: map[string]string{"message": "ok"},
}

func TestMemoryBufferAppendEnqueue(t *testing.T) {
	buf := forward.NewMemoryBuffer(forward.DefaultConfig())
	err := buf.Append(tag, entry)
	if err != nil {
		t.Fatalf("append: %v", err)
	}

	if s := buf.Stat(); s.Stage != 1 {
		t.Fatalf("stat: %v", s)
	}

	buf.Enqueue()
	if s := buf.Stat(); s.Stage != 0 || s.Queue != 1 {
		t.Fatalf("stat: %v", s)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	queue := buf.Dequeue(ctx)

	var chunk *forward.Chunk
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case chunk = <-queue:
		if chunk.Tag != tag {
			t.Fatalf("wrong tag: %#v", chunk)
		}
	}

	buf.Takeback(chunk)
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case chunk = <-queue:
		if chunk.Tag != tag {
			t.Fatalf("wrong tag: %#v", chunk)
		}
	}

	var e forward.Entry
	_, err = e.UnmarshalMsg(chunk.Entries)
	if err != nil {
		t.Fatal(err)
	}
	if n := e.Time.Time(); !n.Equal(entry.Time.Time()) {
		t.Fatalf("time: %v", n)
	}
	if v, ok := e.Record.(map[string]interface{}); ok {
		if v["message"] != "ok" {
			t.Fatalf("record: %#v", e.Record)
		}
	} else {
		t.Fatalf("record: %#v", e.Record)
	}

	err = buf.Purge(chunk)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if s := buf.Stat(); s.Stage != 0 || s.Queue != 0 {
		t.Fatalf("stat: %v", s)
	}

}

func BenchmarkBufferAppend(b *testing.B) {
	cfg := &forward.Config{
		BufferChunkLimit: 0, // disable chunk limit
	}
	b.ReportAllocs()
	b.ResetTimer()
	buf := forward.NewMemoryBuffer(cfg)
	for i := 0; i < b.N; i++ {
		err := buf.Append(tag, entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}
