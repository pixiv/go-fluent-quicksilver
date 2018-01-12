package forward

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Chunk) DecodeMsg(dc *msgp.Reader) (err error) {
	var zbai uint32
	zbai, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if zbai != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zbai}
		return
	}
	z.Tag, err = dc.ReadString()
	if err != nil {
		return
	}
	z.Entries, err = dc.ReadBytes(z.Entries)
	if err != nil {
		return
	}
	var zcmr uint32
	zcmr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if z.Option == nil && zcmr > 0 {
		z.Option = make(map[string]interface{}, zcmr)
	} else if len(z.Option) > 0 {
		for key, _ := range z.Option {
			delete(z.Option, key)
		}
	}
	for zcmr > 0 {
		zcmr--
		var zxvk string
		var zbzg interface{}
		zxvk, err = dc.ReadString()
		if err != nil {
			return
		}
		zbzg, err = dc.ReadIntf()
		if err != nil {
			return
		}
		z.Option[zxvk] = zbzg
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Chunk) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 3
	err = en.Append(0x93)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Tag)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Entries)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Option)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.Option {
		err = en.WriteString(zxvk)
		if err != nil {
			return
		}
		err = en.WriteIntf(zbzg)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Chunk) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 3
	o = append(o, 0x93)
	o = msgp.AppendString(o, z.Tag)
	o = msgp.AppendBytes(o, z.Entries)
	o = msgp.AppendMapHeader(o, uint32(len(z.Option)))
	for zxvk, zbzg := range z.Option {
		o = msgp.AppendString(o, zxvk)
		o, err = msgp.AppendIntf(o, zbzg)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Chunk) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zajw uint32
	zajw, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if zajw != 3 {
		err = msgp.ArrayError{Wanted: 3, Got: zajw}
		return
	}
	z.Tag, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		return
	}
	z.Entries, bts, err = msgp.ReadBytesBytes(bts, z.Entries)
	if err != nil {
		return
	}
	var zwht uint32
	zwht, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if z.Option == nil && zwht > 0 {
		z.Option = make(map[string]interface{}, zwht)
	} else if len(z.Option) > 0 {
		for key, _ := range z.Option {
			delete(z.Option, key)
		}
	}
	for zwht > 0 {
		var zxvk string
		var zbzg interface{}
		zwht--
		zxvk, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		zbzg, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			return
		}
		z.Option[zxvk] = zbzg
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Chunk) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(z.Tag) + msgp.BytesPrefixSize + len(z.Entries) + msgp.MapHeaderSize
	if z.Option != nil {
		for zxvk, zbzg := range z.Option {
			_ = zbzg
			s += msgp.StringPrefixSize + len(zxvk) + msgp.GuessSize(zbzg)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ChunkID) DecodeMsg(dc *msgp.Reader) (err error) {
	err = dc.ReadExactBytes(z[:])
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ChunkID) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes(z[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ChunkID) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, z[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChunkID) UnmarshalMsg(bts []byte) (o []byte, err error) {
	bts, err = msgp.ReadExactBytes(bts, z[:])
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ChunkID) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (16 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Entry) DecodeMsg(dc *msgp.Reader) (err error) {
	var zcua uint32
	zcua, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if zcua != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zcua}
		return
	}
	err = dc.ReadExtension(&z.Time)
	if err != nil {
		return
	}
	z.Record, err = dc.ReadIntf()
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Entry) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return err
	}
	err = en.WriteExtension(&z.Time)
	if err != nil {
		return
	}
	err = en.WriteIntf(z.Record)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Entry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o, err = msgp.AppendExtension(o, &z.Time)
	if err != nil {
		return
	}
	o, err = msgp.AppendIntf(o, z.Record)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Entry) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zxhx uint32
	zxhx, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if zxhx != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zxhx}
		return
	}
	bts, err = msgp.ReadExtensionBytes(bts, &z.Time)
	if err != nil {
		return
	}
	z.Record, bts, err = msgp.ReadIntfBytes(bts)
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Entry) Msgsize() (s int) {
	s = 1 + msgp.ExtensionPrefixSize + z.Time.Len() + msgp.GuessSize(z.Record)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Response) DecodeMsg(dc *msgp.Reader) (err error) {
	var zlqf uint32
	zlqf, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if zlqf != 1 {
		err = msgp.ArrayError{Wanted: 1, Got: zlqf}
		return
	}
	z.Ack, err = dc.ReadString()
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Response) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 1
	err = en.Append(0x91)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Ack)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Response) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 1
	o = append(o, 0x91)
	o = msgp.AppendString(o, z.Ack)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Response) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zdaf uint32
	zdaf, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if zdaf != 1 {
		err = msgp.ArrayError{Wanted: 1, Got: zdaf}
		return
	}
	z.Ack, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Response) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(z.Ack)
	return
}
