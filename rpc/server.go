package rpc

import (
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"net"
)

func Serve(host string) error {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	server := new(stream_transport.Server)
	server.Serve(listener, func(stream stream_transport.IStreamServer) error {
		return nil
	})
	return nil
}
