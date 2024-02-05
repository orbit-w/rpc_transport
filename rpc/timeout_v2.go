package rpc

import (
	"sync"
	"sync/atomic"
	"time"
)

type TimeoutMgr struct {
	max      uint32
	state    atomic.Uint32
	mu       sync.Mutex
	timer    *time.Timer
	itemsMap map[uint32]*Item[uint32, bool]
	queue    expirationQueue[uint32, bool]
	callback func([]uint32)
}

func NewTimeoutMgr(cb func([]uint32)) *TimeoutMgr {
	t := &TimeoutMgr{
		max:      MaxCheck,
		itemsMap: make(map[uint32]*Item[uint32, bool], 0),
		queue:    make(expirationQueue[uint32, bool], 0),
		callback: cb,
	}
	t.state.Store(TTypeRunning)
	t.check()
	return t
}

// Push 如果
func (t *TimeoutMgr) Push(id uint32, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	item, ok := t.get(id)
	if ok {
		t.update(item, ttl)
		return
	}

	t.insert(id, ttl)
}

func (t *TimeoutMgr) Remove(id uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	item, ok := t.get(id)
	if ok {
		head := t.queue.Peek()
		delete(t.itemsMap, id)
		t.queue.Remove(item)
		if item.Equal(head) {
			t.schedule()
		}
	}
}

func (t *TimeoutMgr) OnClose() {
	t.state.CompareAndSwap(TTypeRunning, TTypeStopped)
	t.timer.Stop()
}

func (t *TimeoutMgr) insert(id uint32, ttl time.Duration) {
	item := &Item[uint32, bool]{
		key:       id,
		ttl:       ttl,
		expiresAt: time.Now().Add(ttl),
	}

	head := t.queue.Peek()
	t.itemsMap[id] = item
	t.queue.Enqueue(item)
	if head == nil || !t.queue.Peek().Equal(head) {
		t.schedule()
	}
}

func (t *TimeoutMgr) update(dst *Item[uint32, bool], ttl time.Duration) {
	dst.ttl = ttl
	dst.expiresAt = time.Now().Add(dst.ttl)
	head := t.queue.Peek()
	t.queue.Update(dst)
	if nh := t.queue.Peek(); !nh.Equal(head) || nh.Equal(dst) {
		t.schedule()
	}
}

func (t *TimeoutMgr) get(id uint32) (item *Item[uint32, bool], exist bool) {
	item, exist = t.itemsMap[id]
	return
}

func (t *TimeoutMgr) schedule() {
	if t.state.Load() != TTypeRunning {
		return
	}

	if t.queue.IsEmpty() {
		return
	}

	head := t.queue.Peek()
	if t.timer != nil {
		t.timer.Stop()
	}
	t.timer = time.AfterFunc(time.Until(head.expiresAt), func() {
		t.check()
	})
}

func (t *TimeoutMgr) check() {
	now := time.Now()
	ids := make([]uint32, 0, 1<<3)
	var num uint32
	t.mu.Lock()
	for {
		if t.queue.IsEmpty() {
			break
		}
		front := t.queue[0]
		if front.expiresAt.After(now) {
			break
		}

		ids = append(ids, front.key)
		delete(t.itemsMap, front.key)
		t.queue.Dequeue()
		num++
	}
	t.schedule()
	t.mu.Unlock()
	if num > 0 {
		t.callback(ids)
	}
}
