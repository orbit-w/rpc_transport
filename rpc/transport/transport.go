package transport

import (
	"github.com/orbit-w/golib/bases/packet"
)

/*
   @Author: orbit-w
   @File: transport
   @2023 11月 周日 17:01
*/

type IConn interface {
	Write(data packet.IPacket) error
	Recv() (packet.IPacket, error)
	Close() error
}

type DialOption struct {
	RemoteNodeId      string
	CurrentNodeId     string
	MaxIncomingPacket uint32
	IsBlock           bool
	IsGzip            bool
	DisconnectHandler func(nodeId string)
}

type ConnOption struct {
	MaxIncomingPacket uint32
}
