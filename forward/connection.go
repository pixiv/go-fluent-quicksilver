// Copyright 2018 pixiv Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"net"
	"time"

	"github.com/pkg/errors"
)

// A Connection to a fluentd forward server.
type Connection struct {
	net    string
	addr   string
	dialer net.Dialer
	ctx    context.Context
	config *Config
	conn   net.Conn
}

// NewConnection returns a connection to dial to the named network.
//
// Supported networks are "tcp" and "unix".
func NewConnection(ctx context.Context, network, addr string, config *Config) (*Connection, error) {
	if network != "tcp" && network != "unix" {
		return nil, errors.Errorf("unsupported network: %v", network)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	conn := &Connection{
		net:  network,
		addr: addr,
		dialer: net.Dialer{
			Timeout: config.DialTimeout,
		},
		config: config,
		ctx:    ctx,
	}
	return conn, nil
}

func (c *Connection) Write(p []byte) (n int, err error) {
	err = c.connect()
	if err != nil {
		return
	}

	if c.config.WriteTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	}
	n, err = c.conn.Write(p)
	if err != nil {
		err = errors.Wrap(err, "failed to write")
		c.conn.Close()
		c.conn = nil
	}
	return
}

func (c *Connection) Read(p []byte) (n int, err error) {
	err = c.connect()
	if err != nil {
		return
	}

	if c.config.ReadTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	}
	n, err = c.conn.Read(p)
	if err != nil {
		err = errors.Wrap(err, "failed to read")
		c.conn.Close()
		c.conn = nil
	}
	return
}

func (c *Connection) connect() error {
	if c.conn != nil {
		return nil
	}

	var err error
	c.conn, err = c.dialer.DialContext(c.ctx, c.net, c.addr)
	err = errors.Wrap(err, "failed to connect")
	return err
}

// Close closes the current connection.
func (c *Connection) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

var _ io.ReadWriteCloser = &Connection{}
