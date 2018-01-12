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
	"testing"
	"time"

	"github.com/pixiv/go-fluent-quicksilver/forward"
	"github.com/pixiv/go-fluent-quicksilver/testutil"
)

func TestClientToFluentd(t *testing.T) {
	fluentdConfig := `
    <source>
      @type forward
    </source>
    <match test>
      @type stdout
    </match>
  `

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fluentd, err := testutil.SpawnFluentd(ctx, fluentdConfig, []string{"-q"})
	if err != nil {
		t.Fatal(err)
	}

	config := forward.DefaultConfig()
	config.FlushInterval = time.Millisecond
	config.RetryInterval = time.Millisecond

	client, err := forward.NewClient("tcp", "127.0.0.1:24224", config)
	if err != nil {
		t.Fatal(err)
	}

	client.Post("test", map[string]string{"message": "ok"})

	for {
		if bytes.Contains(fluentd.Bytes(), []byte(`{"message":"ok"}`)) {
			break
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			t.Fatal("fluentd does not respond a message")
		}
	}

	fluentd.Kill()
	<-fluentd.Exit
}
