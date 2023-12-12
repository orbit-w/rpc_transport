package rpc

/*
   @Author: orbit-w
   @File: request
   @2023 12月 周五 23:16
*/

type IRequest interface {
	Id() uint32
	Return()
	Response(in []byte, err error)
	IsAsyncInvoker() bool
	Invoke(in []byte, err error) error
	Done() <-chan IResponse
}

type Request struct {
	seq     uint32
	invoker IAsyncInvoker
	ch      chan IResponse
}

func NewRequest(seq uint32) IRequest {
	r := getRequest()
	r.seq = seq
	return r
}

func NewRequestWithInvoker(seq uint32, _invoker IAsyncInvoker) IRequest {
	r := getRequest()
	r.seq = seq
	r.invoker = _invoker
	return r
}

func (r *Request) Id() uint32 {
	return r.seq
}

func (r *Request) Response(in []byte, err error) {
	rsp := getResponse()
	rsp.in = in
	rsp.err = err
	select {
	case r.ch <- rsp:
	default:
	}
}

func (r *Request) Invoke(in []byte, err error) error {
	return r.invoker.Invoke(in, err)
}

func (r *Request) IsAsyncInvoker() bool {
	return r.invoker == nil
}

func (r *Request) Done() <-chan IResponse {
	return r.ch
}

func (r *Request) Return() {
	r.reset()
	reqPool.Put(r)
}

func (r *Request) reset() {
	r.invoker = nil
	r.seq = 0
}

type IResponse interface {
	In() ([]byte, error)
	Return()
}

type Response struct {
	in  []byte
	err error
}

func (r *Response) In() ([]byte, error) {
	return r.in, NewRpcError(r.err)
}

func (r *Response) Return() {
	r.in = nil
	r.err = nil
	rspPool.Put(r)
}
