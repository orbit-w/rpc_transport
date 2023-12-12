package rpc

import (
	"errors"
	"github.com/orbit-w/orbit-net/core/unbounded"
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
	m       [BucketNum]sync.Map
	cli     *Client
	to      *Timeout[uint32]
}

func (p *Pending) Init(_timeout time.Duration) {
	p.timeout = _timeout
	p.to = NewTimeoutMgr[uint32](_timeout, func(ids []uint32) {
		err := p.cli.input(timeoutListMsg{
			ids: ids,
		})
		if !errors.Is(err, unbounded.ErrCancel) {
			log.Println("handle send timeoutListMsg failed: ", err.Error())
		}
	})
}

func (p *Pending) Push(req IRequest) {
	id := req.Id()
	p.m[id%BucketNum].Store(id, req)
	if req.IsAsyncInvoker() {
		p.to.Push(req.Id())
	}
	return
}

func (p *Pending) Pop(id uint32) (IRequest, bool) {
	v, exist := p.m[id%BucketNum].LoadAndDelete(id)
	if exist {
		req := v.(IRequest)
		if req.IsAsyncInvoker() {
			p.to.Pop(id)
		}
	}
	return nil, false
}

func (p *Pending) OnClose() {
	p.to.OnClose()
}

func (p *Pending) RangeAll(iter func(id uint32)) {
	for i := range p.m {
		p.m[i].Range(func(key, value any) bool {
			iter(key.(uint32))
			return true
		})
	}
}
