package rpc

import (
	"github.com/orbit-w/golib/modules/transport"
	"log"
	"sync"
	"time"
)

/*
   @Author: orbit-w
   @File: pending
   @2023 12月 周六 12:16
*/

const (
	BucketNum = 32
)

type Pending struct {
	timeout time.Duration
	buckets [BucketNum]sync.Map
	cli     *Client
	to      *Timeout
}

func (p *Pending) Init(cli *Client, _timeout time.Duration) {
	p.timeout = _timeout
	p.cli = cli
	p.to = NewTimeoutMgr(func(ids []uint32) {
		if err := p.cli.input(timeoutListMsg{
			ids: ids,
		}); err != nil {
			if !transport.IsCancelError(err) {
				log.Println("[Pending] [Init] handle send timeoutListMsg failed")
			}
		}
	})
}

func (p *Pending) Push(call ICall) {
	id := call.Id()
	bucket := id % BucketNum
	p.buckets[bucket].Store(id, call)
	if call.IsAsyncInvoker() {
		p.to.Push(call.Id(), p.timeout)
	}
	return
}

func (p *Pending) Pop(id uint32) (ICall, bool) {
	bucket := id % BucketNum
	v, exist := p.buckets[bucket].LoadAndDelete(id)
	var call ICall
	if exist {
		call = v.(ICall)
		if call.IsAsyncInvoker() {
			p.to.Remove(id)
		}
	}
	return call, exist
}

func (p *Pending) OnClose() {
	p.to.OnClose()
}

func (p *Pending) RangeAll(iter func(id uint32)) {
	for i := range p.buckets {
		p.buckets[i].Range(func(key, value any) bool {
			iter(key.(uint32))
			return true
		})
	}
}
