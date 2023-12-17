package transport

import (
	"context"
	"fmt"
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"io"
	"log"
	"net"
	"runtime/debug"
	"time"
)

/*
   @Author: orbit-w
   @File: tcp_server
   @2023 11月 周日 21:03
*/

type TcpServer struct {
	authed   bool
	conn     net.Conn
	codec    *TcpCodec
	msgCodec *Codec
	sw       *SenderWrapper
	buf      *ControlBuffer
	rb       *ReceiveBuf
	recvCh   <-chan RecvMsg
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewServerConn(ctx context.Context, _conn net.Conn, ops *ConnOption) IConn {
	if ctx == nil {
		ctx = context.Background()
	}
	cCtx, cancel := context.WithCancel(ctx)
	rb := NewReceiveBuf()
	ts := &TcpServer{
		conn:   _conn,
		codec:  NewTcpCodec(ops.MaxIncomingPacket, false),
		ctx:    cCtx,
		cancel: cancel,
		rb:     rb,
		recvCh: rb.get(),
	}

	sw := NewSender(ts.SendData)
	ts.sw = sw
	ts.buf = NewControlBuffer(ops.MaxIncomingPacket, ts.sw)

	go ts.HandleLoop()
	return ts
}

// Write TcpServer obj does not implicitly call IPacket.Return to return the
// packet to the pool, and the user needs to explicitly call it.
func (ts *TcpServer) Write(data packet.IPacket) (err error) {
	pack := ts.msgCodec.encode(data, TypeMessageRaw)
	err = ts.buf.Set(pack)
	pack.Return()
	return
}

func (ts *TcpServer) Recv() (packet.IPacket, error) {
	select {
	case msg, ok := <-ts.recvCh:
		if !ok {
			return nil, mmrpcs.ErrCanceled
		}
		if msg.err != nil {
			return msg.buf, msg.err
		}
		ts.rb.load()
		return msg.buf, nil
	}
}

func (ts *TcpServer) Close() error {
	log.Println("4444")
	return ts.conn.Close()
}

// SendData implicitly call body.Return
// coding: size<int32> | gzipped<bool> | body<bytes>
func (ts *TcpServer) SendData(body packet.IPacket) error {
	pack := ts.codec.EncodeBody(body)
	defer pack.Return()
	if err := ts.conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	_, err := ts.conn.Write(pack.Data())
	return err
}

func (ts *TcpServer) HandleLoop() {
	header := headPool.Get().(*Buffer)
	buffer := bodyPool.Get().(*Buffer)
	defer func() {
		headPool.Put(header)
		bodyPool.Put(buffer)
	}()

	var (
		err  error
		data packet.IPacket
	)

	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println("stack: ", string(debug.Stack()))
		}
		ts.onClose()
		ts.buf.OnClose()
		if ts.conn != nil {
			_ = ts.conn.Close()
		}
		if err != nil {
			if err == io.EOF || mmrpcs.IsClosedConnError(err) {
				//连接正常断开
			} else {
				log.Println(fmt.Errorf("[TcpServer] tcp_conn disconnected: %s", err.Error()))
			}
		}
	}()

	for {
		data, err = ts.codec.BlockDecodeBody(ts.conn, header.Bytes, buffer.Bytes)
		if err != nil {
			return
		}
		if err = ts.OnData(data); err != nil {
			//TODO: 错误处理？
			return
		}
	}
}

func (ts *TcpServer) onClose() {
	_ = ts.rb.put(RecvMsg{
		err: mmrpcs.ErrCanceled,
	})
}

func (ts *TcpServer) OnData(data packet.IPacket) error {
	defer data.Return()
	for len(data.Remain()) > 0 {
		if bytes, err := data.ReadBytes32(); err == nil {
			reader := packet.Reader(bytes)
			ts.HandleData(reader)
		}
	}
	return nil
}

func (ts *TcpServer) HandleData(in packet.IPacket) {
	mt, data, _ := ts.msgCodec.decode(in)
	switch mt {
	case TypeMessageHeartbeat:
		ack := ts.msgCodec.encode(nil, TypeMessageHeartbeatAck)
		_ = ts.buf.Set(ack)
		ack.Return()
	case TypeMessageHeartbeatAck:
	default:
		_ = ts.rb.put(RecvMsg{
			buf: data,
		})
	}
}
