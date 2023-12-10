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
	reqPool = sync.Pool{New: func() any {
		return new(Request)
	}}

	rspPool = sync.Pool{New: func() any {
		return new(Response)
	}}

	timerPool sync.Pool
)

func getRequest() *Request {
	return reqPool.Get().(*Request)
}

func getResponse() *Response {
	return rspPool.Get().(*Response)
}

func acquireTimer(d time.Duration) *time.Timer {
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

func releaseTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	timerPool.Put(t)
}
