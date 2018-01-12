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

// Package forward provides a forward client for Fluentd.
//
// The implemenation refers to Forward Protocol Specification V1:
// https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1
package forward

import (
	"context"
	"time"
)

// A Config structure is used to configure a forward client.
type Config struct {
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	BufferChunkLimit int
	BufferQueueLimit int
	FlushInterval    time.Duration
	RetryInterval    time.Duration
	MaxRetryInterval time.Duration
}

// DefaultConfig returns a new config with default values.
func DefaultConfig() *Config {
	return &Config{
		DialTimeout:      60 * time.Second,
		ReadTimeout:      60 * time.Second,
		WriteTimeout:     60 * time.Second,
		BufferChunkLimit: 1024 * 1024, // 1MB
		BufferQueueLimit: 512,
		FlushInterval:    time.Second,
		RetryInterval:    time.Second,
		MaxRetryInterval: 60 * time.Second,
	}
}

// A Client that forwards logs to fluentd.
type Client struct {
	conn   *Connection
	output *Output
	cancel context.CancelFunc
}

func NewClient(network, addr string, config *Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := NewConnection(ctx, network, addr, config)
	if err != nil {
		cancel()
		return nil, err
	}

	buf := NewMemoryBuffer(config)
	output := NewOutput(conn, buf, config)

	return &Client{
		conn:   conn,
		output: output,
		cancel: cancel,
	}, nil
}

type WithEventTime interface {
	EventTime() EventTime
}

func (c *Client) Post(tag string, value interface{}) error {
	var t EventTime
	if v, ok := value.(WithEventTime); ok {
		t = v.EventTime()
	} else {
		t = NewEventTime(time.Now())
	}
	return c.output.Append(tag, Entry{t, value})
}

func (c *Client) PostWithTime(tag string, t time.Time, value interface{}) error {
	return c.output.Append(tag, Entry{NewEventTime(t), value})
}

func (c *Client) Close() error {
	c.cancel()
	return nil
}
