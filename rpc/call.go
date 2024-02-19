package rpc

type ICall interface {
	Seq() uint32
	Return()
	Reply(in []byte, err error)
	IsAsyncInvoker() bool
	Invoke(in []byte, err error) error
	Done() <-chan IReply
}

type Call struct {
	seq     uint32
	invoker IAsyncInvoker
	ch      chan IReply
}

func NewCall(seq uint32) ICall {
	r := getCall()
	r.seq = seq
	return r
}

func NewCallWithInvoker(seq uint32, _invoker IAsyncInvoker) ICall {
	r := getCall()
	r.seq = seq
	r.invoker = _invoker
	return r
}

func (r *Call) Seq() uint32 {
	return r.seq
}

func (r *Call) Reply(in []byte, err error) {
	rsp := getReply()
	rsp.in = in
	rsp.err = err
	select {
	case r.ch <- rsp:
	default:
	}
}

func (r *Call) Invoke(in []byte, err error) error {
	return r.invoker.Invoke(in, err)
}

func (r *Call) IsAsyncInvoker() bool {
	return r.invoker != nil
}

func (r *Call) Done() <-chan IReply {
	return r.ch
}

func (r *Call) Return() {
	r.reset()
	callPool.Put(r)
}

func (r *Call) reset() {
	r.invoker = nil
	r.seq = 0
}

type IReply interface {
	In() ([]byte, error)
	Return()
}

type Reply struct {
	in  []byte
	err error
}

func (r *Reply) In() ([]byte, error) {
	return r.in, r.err
}

func (r *Reply) Return() {
	r.in = nil
	r.err = nil
	replyPool.Put(r)
}
