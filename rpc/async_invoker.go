package rpc

import "runtime/debug"

/*
   @Author: orbit-w
   @File: invoker
   @2023 12月 周五 23:13
*/

type IAsyncInvoker interface {
	Invoke(in []byte, err error) error
}

type AsyncInvoker struct {
	ctx     any
	handler func(ctx any, in []byte, err error) error
}

func NewAsyncInvoker(ctx any, cb func(ctx any, in []byte, err error) error) IAsyncInvoker {
	return &AsyncInvoker{ctx: ctx, handler: cb}
}

func (i *AsyncInvoker) Invoke(in []byte, err error) error {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()

	if invokeCB != nil {
		return invokeCB(i.ctx, in, err)
	}
	if i.handler != nil {
		return i.handler(i.ctx, in, err)
	}
	return defineInvokeCB(i.ctx, in, err)
}
