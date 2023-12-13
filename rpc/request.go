package rpc

import (
	"bytes"
	"github.com/orbit-w/golib/bases/packet"
	"io"
)

/*
   @Author: orbit-w
   @File: request
   @2023 12月 周五 23:16
*/

type IRequest interface {
	Pid() int64
	NewReader() io.Reader

	Response(out []byte) error

	// Return To prevent request resource memory leaks,
	// you need to explicitly call the Return method to release it.
	// cannot call Response again later
	Return()
}

type Request struct {
	isResponse bool
	Category   int8 //请求类型： RpcRaw｜RpcCall｜RpcAsyncCall
	seq        uint32
	pid        int64
	buf        []byte
	session    ISession
}

func NewRequest(session ISession, in packet.IPacket) (IRequest, error) {
	d := NewDecoder(in)
	if err := d.Decode(); err != nil {
		return nil, err
	}
	req := reqPool.Get().(*Request)
	req.pid = d.pid
	req.seq = d.seq
	req.Category = d.category
	req.buf = d.buf
	req.session = session
	return req, nil
}

func (r *Request) NewReader() io.Reader {
	return bytes.NewReader(r.buf)
}

func (r *Request) Pid() int64 {
	return r.pid
}

func (r *Request) Response(out []byte) error {
	if r.isResponse {
		return nil
	}
	r.isResponse = true
	return r.session.Send(r.pid, r.seq, r.Category, out)
}

func (r *Request) Return() {
	r.session = nil
	r.buf = nil
	r.seq = 0
	r.pid = 0
	r.Category = 0
	r.isResponse = false
	reqPool.Put(r)
}
