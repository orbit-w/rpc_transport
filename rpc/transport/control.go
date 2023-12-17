package transport

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	err "github.com/orbit-w/orbit-net/core/stream_transport/transport_err"
	"sync"
)

/*
   @Author: orbit-w
   @File: control
   @2023 11月 周日 16:11
*/

type RecvMsg struct {
	buf packet.IPacket
	err error
}

// ReceiveBuf TODO: 资源泄漏？
type ReceiveBuf struct {
	c   chan RecvMsg
	mu  sync.Mutex
	buf []RecvMsg
	err error
}

func NewReceiveBuf() *ReceiveBuf {
	return &ReceiveBuf{
		mu:  sync.Mutex{},
		c:   make(chan RecvMsg, 1),
		buf: make([]RecvMsg, 0),
	}
}

func (rb *ReceiveBuf) OnClose() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.err = mmrpcs.ErrCanceled
	close(rb.c)
}

func (rb *ReceiveBuf) put(r RecvMsg) error {
	rb.mu.Lock()
	if rb.err != nil {
		rb.mu.Unlock()
		return err.ReceiveBufPutErr(rb.err)
	}

	if r.err != nil {
		rb.err = r.err
	}
	if len(rb.buf) == 0 {
		select {
		case rb.c <- r:
			rb.mu.Unlock()
			return nil
		default:
		}
	}
	rb.buf = append(rb.buf, r)
	rb.mu.Unlock()
	return nil
}

func (rb *ReceiveBuf) load() {
	rb.mu.Lock()
	if len(rb.buf) > 0 {
		select {
		case rb.c <- rb.buf[0]:
			rb.buf[0] = RecvMsg{}
			rb.buf = rb.buf[1:]
		default:
		}
	}
	rb.mu.Unlock()
}

func (rb *ReceiveBuf) get() <-chan RecvMsg {
	return rb.c
}
