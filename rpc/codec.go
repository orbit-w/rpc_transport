package rpc

import (
	"fmt"
	"github.com/orbit-w/golib/bases/packet"
)

type Codec struct {
}

func (c Codec) encode(pid int64, seq uint32, category int8, out []byte) packet.IPacket {
	writer := packet.Writer()
	writer.WriteInt64(pid)
	writer.WriteUint32(seq)
	writer.WriteInt8(category)
	if len(out) != 0 {
		writer.Write(out)
	}
	pack := packet.Reader(out)
	return pack
}

type Decoder struct {
	r        packet.IPacket
	category int8
	seq      uint32
	pid      int64
	buf      []byte
}

func NewDecoder(in packet.IPacket) Decoder {
	d := decodersPool.Get().(Decoder)
	d.r = in
	return d
}

func (d Decoder) Decode() error {
	var err error
	d.pid, err = d.r.ReadInt64()
	if err != nil {
		return fmt.Errorf("decode pid failed: %s", err.Error())
	}
	d.seq, err = d.r.ReadUint32()
	if err != nil {
		return fmt.Errorf("decode seq failed: %s", err.Error())
	}
	d.category, err = d.r.ReadInt8()
	if err != nil {
		return fmt.Errorf("decode category failed: %s", err.Error())
	}

	r := d.r.Remain()
	length := len(r)
	dst := make([]byte, length)
	copy(dst, r)
	d.buf = dst
	return nil
}

func (d Decoder) Return() {
	d.buf = nil
	d.pid = 0
	d.seq = 0
	d.category = 0
	if d.r != nil {
		d.r.Return()
		d.r = nil
	}
	decodersPool.Put(d)
}
