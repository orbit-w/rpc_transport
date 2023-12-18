package transport

import (
	"github.com/orbit-w/golib/bases/packet"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
)

type recvMsg struct {
	in  packet.IPacket
	err error
}

func (r recvMsg) Err() error {
	return r.err
}

type iReceiver interface {
	//read blocking read
	read() (in packet.IPacket, err error)
	put(in packet.IPacket, err error)
	onClose(err error)
}

type receiver struct {
	buf *ReceiveBuf[recvMsg]
}

func newReceiver() *receiver {
	return &receiver{
		buf: NewReceiveBuf[recvMsg](),
	}
}

func (r *receiver) read() (in packet.IPacket, err error) {
	select {
	case msg, ok := <-r.buf.get():
		if !ok {
			return nil, mmrpcs.ErrCanceled
		}
		if msg.Err() != nil {
			return msg.in, msg.err
		}
		r.buf.load()
		return msg.in, nil
	}
}

func (r *receiver) put(in packet.IPacket, err error) {
	_ = r.buf.put(recvMsg{
		in:  in,
		err: err,
	})
}

func (r *receiver) onClose(err error) {
	_ = r.buf.put(recvMsg{
		err: err,
	})
}
