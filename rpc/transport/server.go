package transport

import (
	"context"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

/*
   @Author: orbit-w
   @File: server
   @2023 11月 周五 17:04
*/

type Server struct {
	isGzip   bool
	ccu      int32
	host     string
	listener net.Listener
	rw       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc

	handle func(conn IServerConn) error
}

type AcceptorOptions struct {
	MaxIncomingPacket uint32
	IsGzip            bool
}

func (ins *Server) Serve(listener net.Listener, _handle func(conn IServerConn) error, ops ...AcceptorOptions) {
	op := parseAndWrapOP(ops...)
	NewBodyPool(op.MaxIncomingPacket)
	ctx, cancel := context.WithCancel(context.Background())
	ins.rw = sync.RWMutex{}
	ins.host = ""
	ins.isGzip = op.IsGzip
	ins.ctx = ctx
	ins.cancel = cancel
	ins.handle = _handle
	ins.listener = listener
	go ins.acceptLoop()
}

func (ins *Server) acceptLoop() {
	for {
		conn, err := ins.listener.Accept()
		if err != nil {
			select {
			case <-ins.ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		ins.handleConn(NewServerConn(ins.ctx, conn, &ConnOption{
			MaxIncomingPacket: MaxIncomingPacket,
		}))
	}
}

func (ins *Server) handleConn(conn IServerConn) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
				log.Println("stack", string(debug.Stack()))
			}
			_ = conn.Close()
		}()

		if appErr := ins.handle(conn); appErr != nil {
			//TODO:
		}
	}()
}

func parseAndWrapOP(ops ...AcceptorOptions) AcceptorOptions {
	var op AcceptorOptions
	if len(ops) > 0 {
		op = ops[0]
	}
	if op.MaxIncomingPacket == 0 {
		op.MaxIncomingPacket = MaxIncomingPacket
	}
	return op
}
