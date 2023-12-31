package rpc

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	TTypeNone = iota
	TTypeRunning
	TTypeStopped

	MaxCheck = 20
)

type Timeout struct {
	max      uint32
	state    atomic.Uint32
	mu       sync.Mutex
	itemsMap map[uint32]*Item[uint32, bool]
	queue    expirationQueue[uint32, bool]
	callback func([]uint32)
}

func NewTimeoutMgr(cb func([]uint32)) *Timeout {
	t := &Timeout{
		max:      MaxCheck,
		itemsMap: make(map[uint32]*Item[uint32, bool], 0),
		queue:    make(expirationQueue[uint32, bool], 0),
		callback: cb,
	}
	t.state.Store(TTypeRunning)
	t.schedule()
	return t
}

func (t *Timeout) Push(id uint32, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	item, ok := t.get(id)
	if ok {
		item.ttl = ttl
		item.expiresAt = time.Now().Add(item.ttl)
		t.queue.update(item)
		return
	}

	item = &Item[uint32, bool]{
		key:       id,
		ttl:       ttl,
		expiresAt: time.Now().Add(ttl),
	}
	t.itemsMap[id] = item
	t.queue.push(item)
}

func (t *Timeout) Remove(id uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	item, ok := t.get(id)
	if ok {
		delete(t.itemsMap, id)
		t.queue.remove(item)
	}
}

func (t *Timeout) OnClose() {
	t.state.CompareAndSwap(TTypeRunning, TTypeStopped)
}

func (t *Timeout) get(id uint32) (item *Item[uint32, bool], exist bool) {
	item, exist = t.itemsMap[id]
	return
}

func (t *Timeout) schedule() {
	if t.state.Load() == TTypeRunning {
		time.AfterFunc(time.Second, func() {
			t.check()
		})
	}
}

func (t *Timeout) check() {
	t.schedule()
	now := time.Now()
	ids := make([]uint32, 0, 1<<3)
	var num uint32
	t.mu.Lock()
	for {
		if t.queue.isEmpty() {
			break
		}
		front := t.queue[0]
		if front.expiresAt.After(now) {
			break
		}
		delete(t.itemsMap, front.key)
		t.queue.remove(front)
		ids = append(ids, front.key)
		num++
	}
	t.mu.Unlock()
	if num > 0 {
		t.callback(ids)
	}
}
