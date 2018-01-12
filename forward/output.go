package forward

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
)

type Output struct {
	rw     io.ReadWriteCloser
	buffer Buffer
	config *Config
}

func NewOutput(rw io.ReadWriteCloser, buf Buffer, config *Config) *Output {
	output := &Output{
		rw:     rw,
		buffer: buf,
		config: config,
	}
	return output
}

func (o *Output) Run(ctx context.Context) error {
	go o.writeLoop(ctx)
	go o.flushLoop(ctx)
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (o *Output) Append(tag string, entry Entry) error {
	return o.buffer.Append(tag, entry)
}

func (o *Output) writeLoop(ctx context.Context) {
	buf := make([]byte, o.config.BufferChunkLimit)
	retryInterval := o.config.RetryInterval
	dequeue := o.buffer.Dequeue(ctx)
	for {
		var chunk *Chunk
		select {
		case <-ctx.Done():
			return
		case chunk = <-dequeue:
		}
		err := o.write(chunk, buf)
		if err != nil {
			o.buffer.Takeback(chunk)

			// wait for retry
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
			}
			retryInterval = retryInterval * 2
			if retryInterval > o.config.MaxRetryInterval {
				retryInterval = o.config.MaxRetryInterval
			}
			continue
		}
		retryInterval = o.config.RetryInterval
	}
}

func (o *Output) write(chunk *Chunk, buf []byte) error {
	var err error
	buf, err = chunk.MarshalMsg(buf[:0])
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}
	_, err = o.rw.Write(buf)
	if err != nil {
		return err
	}
	o.buffer.Purge(chunk)
	return nil
}

func (o *Output) flushLoop(ctx context.Context) {
	t := time.NewTicker(o.config.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		o.buffer.Enqueue()
	}
}
