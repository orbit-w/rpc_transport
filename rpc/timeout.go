package rpc

import (
	"github.com/huandu/skiplist"
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
	timeout  int64
	skipList *skiplist.SkipList
	callback func([]uint32)
}

func NewTimeoutMgr(timeout time.Duration, cb func([]uint32)) *Timeout {
	t := &Timeout{
		mu:       sync.Mutex{},
		max:      MaxCheck,
		timeout:  int64(timeout / time.Second),
		skipList: skiplist.New(skiplist.Uint32),
		callback: cb,
	}
	t.state.Store(TTypeRunning)
	//t.schedule()
	return t
}

func (t *Timeout) Push(id uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.skipList.Set(id, time.Now().Unix()+t.timeout)
}

func (t *Timeout) Pop(id uint32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.skipList.Remove(id)
}

func (t *Timeout) OnClose() {
	t.state.CompareAndSwap(TTypeRunning, TTypeStopped)
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
	now := time.Now().Unix()
	ids := make([]uint32, 0, 1<<3)
	var num uint32
	t.mu.Lock()
	t.skipList.Front()

	for {
		head := t.skipList.Front()
		if head == nil || head.Value.(int64) > max {
			break
		}

		s := head.Value.(int64)

		h.h.Pop()
		v := head.Value
		key := v.GetKey()
		delete(h.items, key)

		if !iter(key, v.Value) {
			break
		}
	}

	t.skipList.RemoveFront()
	t.heapList.PopByScore(now, func(k T, _ int8) bool {
		if t.max > 0 && num >= t.max {
			return false
		}
		ids = append(ids, k)
		num++
		return true
	})
	t.mu.Unlock()
	if num > 0 {
		t.callback(ids)
	}
}
