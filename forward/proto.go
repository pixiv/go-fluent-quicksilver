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
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate $GOPATH/bin/msgp
//msgp:tuple Chunk Entry Response

// A Chunk of entries.
type Chunk struct {
	Tag     string
	Entries []byte
	Option  map[string]interface{}
}

// A Entry in msgpack stream.
type Entry struct {
	Time   EventTime `msg:",extension"`
	Record interface{}
}

// A Response from forward server for a message.
type Response struct {
	Ack string `msg:"ack"`
}

// EventTime is a msgpack extension format to carry
// nanosecond precision of time.
type EventTime struct {
	sec, nsec uint32
}

// EventTimeExtensionType is 0.
const EventTimeExtensionType = 0

// NewEventTime returns a EventTime of time t.
func NewEventTime(t time.Time) EventTime {
	return EventTime{
		sec:  uint32(t.Unix()),
		nsec: uint32(t.Nanosecond()),
	}
}

func newEventTimeExt() msgp.Extension {
	return new(EventTime)
}

// ExtensionType implements msgp.Extension.
func (t *EventTime) ExtensionType() int8 {
	return 0
}

// Len implements msgp.Extension.
func (t *EventTime) Len() int {
	return 8
}

// MarshalBinaryTo implements msgp.Extension.
func (t *EventTime) MarshalBinaryTo(p []byte) error {
	_ = p[7]
	p[0] = byte(t.sec >> 24)
	p[1] = byte(t.sec >> 16)
	p[2] = byte(t.sec >> 8)
	p[3] = byte(t.sec)
	p[4] = byte(t.nsec >> 24)
	p[5] = byte(t.nsec >> 16)
	p[6] = byte(t.nsec >> 8)
	p[7] = byte(t.nsec)
	return nil
}

// UnmarshalBinary implements msgp.Extension.
func (t *EventTime) UnmarshalBinary(p []byte) error {
	t.sec = binary.BigEndian.Uint32(p[0:])
	t.nsec = binary.BigEndian.Uint32(p[4:])
	return nil
}

// Time returns a instant of Time.
func (t *EventTime) Time() time.Time {
	return time.Unix(int64(t.sec), int64(t.nsec))
}

func init() {
	msgp.RegisterExtension(EventTimeExtensionType, newEventTimeExt)
}

// ChunkID is a unique ID for a chunk
type ChunkID [16]byte

// NewChunkID generates and returns a unique ChunkID.
func NewChunkID() ChunkID {
	var id ChunkID
	rand.Read(id[0:])
	return id
}

func (id ChunkID) String() string {
	return base64.StdEncoding.EncodeToString(id[0:])
}

// DecodeChunkID decodes base64 encoded chunk ID string s.
// If the given string is invalid, it returns an error.
func DecodeChunkID(s string) (ChunkID, error) {
	var id ChunkID
	b, err := base64.StdEncoding.DecodeString(s)
	copy(id[0:], b)
	return id, err
}
