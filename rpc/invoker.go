package rpc

/*
   @Author: orbit-w
   @File: invoker
   @2023 12月 周五 23:13
*/

type IInvoker interface {
	Invoke(in []byte, err error) error
}

type Invoker struct {
	ctx     any
	handler func(ctx any, in []byte, err error) error
}

func (i *Invoker) Invoke(in []byte, err error) error {
	return i.handler(i.ctx, in, err)
}
