package rpc

import (
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"github.com/orbit-w/orbit-net/core/stream_transport/metadata"
	"log"
	"net"
)

func Serve(host string, rh RequestHandle) error {
	if rh == nil {
		setTestHandle()
	} else {
		gRequestHandle = rh
	}

	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	server := new(stream_transport.Server)
	server.Serve(listener, func(stream stream_transport.IStreamServer) error {
		ctx := stream.Context()
		md, _ := metadata.FromMetaContext(ctx)
		nodeId, _ := md.GetValue("nodeId")
		log.Println("Connection established successfully, client nodeId: ", nodeId)
		NewConn(stream)
		return nil
	})
	return nil
}

type RequestHandle func(req IRequest) error

var gRequestHandle RequestHandle

func setTestHandle() {
	gRequestHandle = func(req IRequest) error {
		_ = req.Response([]byte{1})
		return nil
	}
}
