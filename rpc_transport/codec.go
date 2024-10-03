package rpc_transport

import (
	"encoding/binary"
	"errors"
	"github.com/orbit-w/meteor/modules/net/packet"
)

type Codec struct {
}

func (c Codec) Encode(seq uint32, category int8, out []byte) packet.IPacket {
	writer := packet.WriterP(4 + 1 + len(out))
	writer.WriteUint32(seq)
	writer.WriteInt8(category)
	if len(out) != 0 {
		writer.Write(out)
	}
	return writer
}

type Decoder struct {
	category int8
	seq      uint32
	pid      int64
	buf      []byte
}

func NewDecoder() *Decoder {
	d := decodersPool.Get().(*Decoder)
	return d
}

func (d *Decoder) Decode(in []byte) error {
	if len(in) < (4 + 1) {
		return errors.New("decode failed")
	}
	var off int
	d.seq = binary.BigEndian.Uint32(in[off : off+4])
	off += 4
	d.category = int8(in[off])
	off += 1
	d.buf = in[off:]
	return nil
}

func (d *Decoder) Return() {
	d.buf = nil
	d.pid = 0
	d.seq = 0
	d.category = 0
	decodersPool.Put(d)
}
