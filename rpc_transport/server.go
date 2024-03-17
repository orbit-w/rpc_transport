package rpc_transport

import (
	"github.com/orbit-w/golib/core/transport"
	"net"
)

type RpcServer struct {
	ts *transport.Server
}

func (s *RpcServer) Serve(host string, rh RequestHandle) error {
	if rh == nil {
		setTestHandle()
	} else {
		gRequestHandle = rh
	}

	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	server := new(transport.Server)
	server.Serve(listener, func(conn transport.IServerConn) error {
		//ctx := stream.Context()
		//md, _ := metadata.FromMetaContext(ctx)
		//nodeId, _ := md.GetValue("nodeId")
		//log.Println("Connection established successfully, client nodeId: ", nodeId)
		NewConn(conn)
		return nil
	})
	s.ts = server
	return nil
}

func (s *RpcServer) Stop() error {
	if s.ts != nil {
		return s.ts.Stop()
	}
	return nil
}

// RequestHandle To avoid IRequest resource leakage,
// IRequest requires the receiver to actively call Return to return it to the pool
type RequestHandle func(req IRequest) error

var gRequestHandle RequestHandle

func SetReqHandle(h RequestHandle) {
	gRequestHandle = h
}

func setTestHandle() {
	gRequestHandle = func(req IRequest) error {
		//log.Println("receive request")
		_ = req.Response([]byte{1})
		req.Return()
		return nil
	}
}
