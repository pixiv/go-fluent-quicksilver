package forward_test

import (
	"bytes"
	"context"
	"errors"
	"io"
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

func TestOutput(t *testing.T) {
	success := make(chan bool)
	output := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, output},
	}

	config := forward.DefaultConfig()
	config.FlushInterval = time.Millisecond
	buf := forward.NewMemoryBuffer(config)
	out := forward.NewOutput(rw, buf, config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go out.Run(ctx)

	err := out.Append(tag, entry)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case success <- true:
	}

	var p []byte
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case p = <-output:
	}

	var chunk forward.Chunk
	_, err = chunk.UnmarshalMsg(p)
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Tag != tag {
		t.Fatalf("chunk: %v", chunk.Tag)
	}
}

func TestOutputRetry(t *testing.T) {
	success := make(chan bool)
	output := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, output},
	}

	config := forward.DefaultConfig()
	config.FlushInterval = time.Millisecond
	config.RetryInterval = time.Millisecond
	config.MaxRetryInterval = 10 * time.Millisecond

	buf := forward.NewMemoryBuffer(config)
	out := forward.NewOutput(rw, buf, config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go out.Run(ctx)

	err := out.Append(tag, entry)
	if err != nil {
		t.Fatal(err)
	}

	// will fail 5 times
	start := time.Now()
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case success <- false:
			ms := float64(time.Now().Sub(start).Nanoseconds()) / 1000.0 / 1000.0
			t.Logf("[%7.4f ms] fail: %v", ms, i)
		}
	}

	// will success
	select {
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	case success <- true:
		ms := float64(time.Now().Sub(start).Nanoseconds()) / 1000.0 / 1000.0
		t.Logf("[%7.4f ms] success", ms)
	}

	<-output
}

func BenchmarkOutput(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	success := make(chan bool)
	output := make(chan []byte)
	rw := &rw{
		r: bytes.NewBuffer(nil),
		w: &testWriter{success, output},
	}

	b.ReportAllocs()
	b.ResetTimer()

	config := forward.DefaultConfig()
	buf := forward.NewMemoryBuffer(config)
	out := forward.NewOutput(rw, buf, config)
	go out.Run(ctx)

	for i := 0; i < b.N; i++ {
		err := out.Append(tag, entry)
		if err != nil {
			b.Fatal(err)
		}
		buf.Enqueue()
		success <- true
		<-output
	}
}
