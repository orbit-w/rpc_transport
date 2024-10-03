package rpc_transport

import (
	"fmt"
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
	reader := packet.ReaderP(in)
	defer packet.Return(reader)

	var err error

	d.seq, err = reader.ReadUint32()
	if err != nil {
		return fmt.Errorf("decode seq failed: %s", err.Error())
	}

	d.category, err = reader.ReadInt8()
	if err != nil {
		return fmt.Errorf("decode category failed: %s", err.Error())
	}

	d.buf = reader.CopyRemain()
	return nil
}

func (d *Decoder) Return() {
	d.buf = nil
	d.pid = 0
	d.seq = 0
	d.category = 0
	decodersPool.Put(d)
}
