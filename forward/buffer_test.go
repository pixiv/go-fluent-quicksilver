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
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/pixiv/go-fluent-quicksilver/forward"
)

type rw struct {
	r io.Reader
	w io.Writer
}

func (rw *rw) Read(p []byte) (int, error) {
	return rw.r.Read(p)
}

func (rw *rw) Write(p []byte) (int, error) {
	return rw.w.Write(p)
}

func (rw *rw) Close() error {
	return nil
}

type testWriter struct {
	success chan bool
	output  chan []byte
}

func (w *testWriter) Write(p []byte) (int, error) {
	success := <-w.success
	if success {
		w.output <- p
		return len(p), nil
	}
	return 0, errors.New("not success")
}

var tag = "test"
var entry = forward.Entry{
	Time:   forward.NewEventTime(time.Now()),
	Record: map[string]string{"message": "ok"},
}

func TestBufferAppendEnqueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	success := make(chan bool)
	output := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, output},
	}

	buf := forward.NewBuffer(ctx, rw, forward.DefaultConfig())
	err := buf.Append(tag, entry)
	if err != nil {
		t.Fatal(err)
	}
	if s := buf.Stat(); s.Stage != 1 {
		t.Fatalf("stat: %v", s)
	}

	buf.Enqueue(tag)
	if s := buf.Stat(); s.Stage != 0 || s.Queue != 1 {
		t.Fatalf("stat: %v", s)
	}

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case success <- true:
	}

	var out []byte
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case out = <-output:
	}

	var c forward.Chunk
	_, err = c.UnmarshalMsg(out)
	if err != nil {
		t.Fatal(err)
	}
	if c.Tag != tag {
		t.Fatalf("tag: %v", c.Tag)
	}

	var e forward.Entry
	_, err = e.UnmarshalMsg(c.Entries)
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
}

func TestBufferRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	success := make(chan bool)
	output := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, output},
	}

	config := forward.DefaultConfig()
	config.RetryInterval = time.Millisecond
	config.MaxRetryInterval = 5 * time.Millisecond

	buf := forward.NewBuffer(ctx, rw, config)

	err := buf.Append(tag, entry)
	if err != nil {
		t.Fatal(err)
	}

	// write will fail 5 times
	for i := 0; i < 5; i++ {
		buf.Enqueue(tag)
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case success <- false:
		}
	}

	// write will success
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case success <- true:
	}

	<-output
}

func BenchmarkBufferAppend(b *testing.B) {
	cfg := &forward.Config{
		BufferChunkLimit: 0, // disable chunk limit
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	buf := forward.NewBuffer(ctx, rw, cfg)
	for i := 0; i < b.N; i++ {
		err := buf.Append(tag, entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBufferAppendEnqueue(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	buf := forward.NewBuffer(ctx, rw, forward.DefaultConfig())
	for i := 0; i < b.N; i++ {
		err := buf.Append(tag, entry)
		if err != nil {
			b.Fatal(err)
		}
		buf.Enqueue(tag)
	}
}

func BenchmarkBufferWrite(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	success := make(chan bool)
	out := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, out},
	}

	b.ReportAllocs()
	b.ResetTimer()
	buf := forward.NewBuffer(ctx, rw, forward.DefaultConfig())
	for i := 0; i < b.N; i++ {
		err := buf.Append(tag, entry)
		if err != nil {
			b.Fatal(err)
		}
		buf.Enqueue(tag)
		success <- true
		<-out
	}
}
