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
	"io"
	"net"
	"testing"

	"github.com/pixiv/go-fluent-quicksilver/forward"
)

func testEchoServer(t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Log(err)
				continue
			}
			go func(c net.Conn) {
				io.Copy(c, c)
				c.Close()
			}(conn)
		}
	}(l)
	return l
}

func TestConnectionReadWrite(t *testing.T) {
	echo := testEchoServer(t)
	defer echo.Close()
	net, addr := echo.Addr().Network(), echo.Addr().String()

	conn, err := forward.NewConnection(nil, net, addr, forward.DefaultConfig())
	if err != nil {
		t.Fatalf("NewConnection: %+v", err)
	}

	msg := []byte("hello")

	n, err := conn.Write(msg)
	if err != nil {
		t.Fatalf("Write: %+v", err)
	}
	if n != len(msg) {
		t.Fatalf("Write: invalid length: %v want %v", n, len(msg))
	}

	resp := make([]byte, len(msg))
	_, err = conn.Read(resp)
	if err != nil {
		t.Fatalf("Read: %#v", err)
	}
	if !bytes.Equal(msg, resp) {
		t.Fatalf("Read: response: %#v", resp)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}
