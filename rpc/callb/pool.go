package callb

import (
	"sync"
	"time"
)

var (
	callPool = sync.Pool{New: func() any {
		return new(Call)
	}}

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
