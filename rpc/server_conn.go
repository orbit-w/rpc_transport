package rpc

import (
	"github.com/orbit-w/orbit-net/core/stream_transport"
)

type Conn struct {
	stream stream_transport.IStreamServer
}

//func (c *Conn) LoopRead() {
//	for {
//		in, err := c.stream.Recv()
//		if err != nil {
//			if IsCancelError(err) {
//				break
//			}
//			log.Println("conn read stream failed: ", err.Error())
//			break
//		}
//
//	}
//}
