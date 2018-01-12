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
	"testing"

	"github.com/pixiv/go-fluent-quicksilver/forward"
)

func TestChunkID(t *testing.T) {
	id := forward.NewChunkID()

	s := id.String()
	decoded, err := forward.DecodeChunkID(s)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != id {
		t.Fatalf("invalid ID: %v", s)
	}
}

func BenchmarkNewChunkID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = forward.NewChunkID().String()
	}
}
