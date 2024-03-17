package rpc_transport

import (
	"fmt"
	"github.com/orbit-w/golib/bases/packet"
)

type Codec struct {
}

func (c Codec) encode(seq uint32, category int8, out []byte) packet.IPacket {
	writer := packet.Writer()
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

func (d *Decoder) Decode(in packet.IPacket) error {
	defer in.Return()
	var err error
	d.seq, err = in.ReadUint32()
	if err != nil {
		return fmt.Errorf("decode seq failed: %s", err.Error())
	}
	d.category, err = in.ReadInt8()
	if err != nil {
		return fmt.Errorf("decode category failed: %s", err.Error())
	}

	r := in.Remain()
	length := len(r)
	dst := make([]byte, length)
	copy(dst, r)
	d.buf = dst
	return nil
}

func (d *Decoder) Return() {
	d.buf = nil
	d.pid = 0
	d.seq = 0
	d.category = 0
	decodersPool.Put(d)
}
