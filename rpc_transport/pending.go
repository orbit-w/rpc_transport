package rpc_transport

import (
	"github.com/orbit-w/meteor/modules/net/transport"
	"github.com/orbit-w/rpc_transport/timeout"
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
	to      *timeout.Timeout
}

func (p *Pending) Init(cli *Client, _timeout time.Duration) {
	p.timeout = _timeout
	p.cli = cli
	p.to = timeout.NewTimeout(func(ids []uint32) {
		if err := p.cli.input(timeoutListMsg{
			ids: ids,
		}); err != nil {
			if !transport.IsCancelError(err) {
				cli.log.Error("[Pending] [Init] handle send timeoutListMsg failed")
			}
		}
	})
}

func (p *Pending) Push(call ICall) {
	id := call.Seq()
	bucket := id % BucketNum
	p.buckets[bucket].Store(id, call)
	if call.IsAsyncInvoker() {
		p.to.Insert(call.Seq(), p.timeout)
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
