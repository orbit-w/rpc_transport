package rpc

import (
	"github.com/orbit-w/golib/bases/container/heap_list"
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

type Timeout[T comparable] struct {
	max      uint32
	state    atomic.Uint32
	mu       sync.Mutex
	timeout  time.Duration
	heapList *heap_list.HeapList[T, int8, int64]
	callback func([]T)
}

func NewTimeoutMgr[T comparable](timeout time.Duration, cb func([]T)) *Timeout[T] {
	t := &Timeout[T]{
		max:      MaxCheck,
		timeout:  timeout,
		heapList: heap_list.New[T, int8, int64](),
		callback: cb,
	}
	t.state.Store(TTypeRunning)
	t.schedule()
	return t
}

func (t *Timeout[T]) Push(id T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.heapList.Push(id, 0, time.Now().Add(t.timeout).Unix())
}

func (t *Timeout[T]) Pop(id T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.heapList.Delete(id)
}

func (t *Timeout[T]) OnClose() {
	t.state.CompareAndSwap(TTypeRunning, TTypeStopped)
}

func (t *Timeout[T]) schedule() {
	if t.state.Load() == TTypeRunning {
		time.AfterFunc(time.Second, func() {
			t.check()
		})
	}
}

func (t *Timeout[T]) check() {
	t.schedule()
	now := time.Now().Unix()
	ids := make([]T, 0, 1<<3)
	var num uint32
	t.mu.Lock()
	t.heapList.PopByScore(now, func(k T, _ int8) bool {
		if t.max > 0 && num >= t.max {
			return false
		}
		ids = append(ids, k)
		num++
		return true
	})
	t.mu.Unlock()

	t.callback(ids)
}
