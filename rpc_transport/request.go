package rpc_transport

import (
	"github.com/orbit-w/meteor/modules/net/packet"
)

/*
   @Author: orbit-w
   @File: request
   @2023 12月 周五 23:16
*/

// IRequest To avoid IRequest resource leakage,
// IRequest requires the receiver to call Return() to return it to the pool
type IRequest interface {
	NewReader() packet.IPacket
	Data() []byte
	Category() int8
	Response(out []byte) error
	IgnoreRsp()

	// Return To prevent request resource memory leaks,
	// you need to explicitly call the Return method to release it.
	// cannot call Response again later
	Return()
}

type Request struct {
	isResponse bool
	ignoreRsp  bool
	category   int8 //请求类型： RpcRaw｜RpcCall｜RpcAsyncCall
	seq        uint32
	buf        []byte
	session    ISession
}

func NewRequest(session ISession, in []byte) (IRequest, error) {
	d := NewDecoder()
	defer d.Return()
	if err := d.Decode(in); err != nil {
		return nil, err
	}
	req := reqPool.Get().(*Request)
	req.seq = d.seq
	req.category = d.category
	req.buf = d.buf
	req.session = session
	return req, nil
}

func (r *Request) NewReader() packet.IPacket {
	return packet.ReaderP(r.buf)
}

func (r *Request) Data() []byte {
	return r.buf
}

func (r *Request) Category() int8 {
	return r.category
}

func (r *Request) IgnoreRsp() {
	r.ignoreRsp = true
}

func (r *Request) Response(out []byte) error {
	if r.isResponse || r.ignoreRsp {
		return nil
	}

	r.isResponse = true
	return r.session.Send(r.seq, r.category, out)
}

func (r *Request) Return() {
	r.session = nil
	r.buf = nil
	r.seq = 0
	r.category = 0
	r.isResponse = false
	r.ignoreRsp = false
	reqPool.Put(r)
}
