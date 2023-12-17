package transport

import "github.com/orbit-w/golib/bases/packet"

/*
   @Author: orbit-w
   @File: codec
   @2023 12月 周六 20:41
*/

type Codec struct {
}

func (c *Codec) encode(data packet.IPacket, mt int8) packet.IPacket {
	writer := packet.Writer()
	writer.WriteInt8(mt)
	if data != nil && len(data.Remain()) > 0 {
		writer.Write(data.Remain())
	}
	return writer
}

func (c *Codec) decode(data packet.IPacket) (mt int8, writer packet.IPacket, err error) {
	defer data.Return()
	mt, err = data.ReadInt8()
	if err != nil {
		return
	}

	if len(data.Remain()) > 0 {
		writer = packet.Writer()
		writer.Write(data.Remain())
	}
	return
}
