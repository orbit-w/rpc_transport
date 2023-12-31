package rpc

import (
	"sync"
	"time"
)

/*
	@Author: orbit-w
	@File: pool
	@2023 12月 周六 11:20
*/

var (
	decodersPool = sync.Pool{New: func() any {
		return new(Decoder)
	}}

	reqPool = sync.Pool{New: func() any {
		return new(Request)
	}}

	callPool = sync.Pool{}

	replyPool = sync.Pool{New: func() any {
		return new(Reply)
	}}

	timerPool sync.Pool
)

func getCall() *Call {
	v := callPool.Get()
	if v == nil {
		return &Call{
			ch: make(chan IReply, 1),
		}
	}
	return v.(*Call)
}

func getReply() *Reply {
	return replyPool.Get().(*Reply)
}

func AcquireTimer(d time.Duration) *time.Timer {
	v := timerPool.Get()
	if v == nil {
		return time.NewTimer(d)
	}
	t := v.(*time.Timer)
	if t.Reset(d) {
		t = time.NewTimer(d)
	}
	return t
}

func ReleaseTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	timerPool.Put(t)
}
