package rpc

import (
	"github.com/orbit-w/golib/bases/packet"
	"log"
)

/*
   @Author: orbit-w
   @File: client_codec
   @2023 12月 周日 22:25
*/

/*
codec:

  - pid int64 - | - seq uint32 - | - category int8 - | - data []byte -
*/
func (c *Client) encode(pid int64, seq uint32, category int8, out []byte) packet.IPacket {
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

func (c *Client) decode(in packet.IPacket) (msg message, err error) {
	defer in.Return()
	msg.pid, err = in.ReadInt64()
	if err != nil {
		log.Println("decode pid failed: ", err.Error())
		return
	}
	msg.seq, err = in.ReadUint32()
	if err != nil {
		log.Println("decode seq failed: ", err.Error())
		return
	}
	msg.category, err = in.ReadInt8()
	if err != nil {
		log.Println("decode category failed: ", err.Error())
		return
	}

	r := in.Remain()
	length := len(r)
	dst := make([]byte, length)
	copy(dst, r)
	msg.reply = dst
	return
}
