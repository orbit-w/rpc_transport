package rpc

import (
	"github.com/orbit-w/golib/bases/container/heap_list"
	"sync"
	"time"
)

/*
   @Author: orbit-w
   @File: pending
   @2023 12月 周六 12:16
*/

type Pending struct {
	mu       sync.Mutex
	timeout  time.Duration
	hmu      sync.Mutex
	m        map[uint32]IRequest
	heapList *heap_list.HeapList[uint32, int8, int64]
}

func (p *Pending) Init(_timeout time.Duration) {
	p.mu = sync.Mutex{}
	p.timeout = _timeout
	p.m = make(map[uint32]IRequest)
	p.heapList = heap_list.New[uint32, int8, int64]()
}

func (p *Pending) Push(req IRequest) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[req.Id()] = req
	if req.IsInvoker() {
		p.heapList.Push(req.Id(), 0, time.Now().Add(p.timeout).Unix())
	}
	return
}

func (p *Pending) Pop(id uint32) (IRequest, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	req, ok := p.m[id]
	if ok {
		delete(p.m, id)
		if req.IsInvoker() {
			p.heapList.Delete(id)
		}
	}
	return nil, false
}

func (p *Pending) RangeTimeout(iter func(req IRequest)) {
	now := time.Now().Unix()
	p.mu.Lock()
	defer p.mu.Unlock()
	ids := make([]uint32, 0, 1<<3)
	p.heapList.PopByScore(now, func(k uint32, _ int8) bool {
		ids = append(ids, k)
		return true
	})

	for i := range ids {
		id := ids[i]
		req, ok := p.m[id]
		if ok {
			delete(p.m, id)
			iter(req)
		}
	}
}
