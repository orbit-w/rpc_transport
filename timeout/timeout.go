package timeout

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	TTypeRunning = iota + 1
	TTypeStopped

	BucketNum = 8
)

type Timeout struct {
	state   atomic.Uint32
	buckets [BucketNum]*Bucket
}

func NewTimeout(cb func([]uint32)) *Timeout {
	t := &Timeout{
		buckets: [BucketNum]*Bucket{},
		state:   atomic.Uint32{},
	}
	t.state.Store(TTypeRunning)
	for i := 0; i < BucketNum; i++ {
		t.buckets[i] = NewBucket(t, cb)
	}
	return t
}

func (t *Timeout) Insert(id uint32, ttl time.Duration) {
	index := id % BucketNum
	bucket := t.buckets[index]
	bucket.Insert(id, ttl)
}

func (t *Timeout) Remove(id uint32) {
	index := id % BucketNum
	bucket := t.buckets[index]
	bucket.Remove(id)
}

func (t *Timeout) OnClose() {
	t.state.CompareAndSwap(TTypeRunning, TTypeStopped)
	for i := range t.buckets {
		b := t.buckets[i]
		b.OnClose()
	}
}

func (t *Timeout) State() uint32 {
	return t.state.Load()
}

type Bucket struct {
	rw       sync.RWMutex
	mgr      *Timeout
	timer    *time.Timer
	queue    expirationQueue[uint32, bool]
	itemsMap map[uint32]*Item[uint32, bool]
	callback func([]uint32)
}

func NewBucket(mgr *Timeout, cb func([]uint32)) *Bucket {
	b := &Bucket{
		rw:       sync.RWMutex{},
		mgr:      mgr,
		itemsMap: make(map[uint32]*Item[uint32, bool], 0),
		queue:    newExpirationQueue[uint32, bool](),
		callback: cb,
	}
	b.schedule()
	return b
}

func (b *Bucket) Insert(id uint32, ttl time.Duration) {
	b.rw.Lock()
	defer b.rw.Unlock()
	item := &Item[uint32, bool]{
		key:       id,
		ttl:       ttl,
		expiresAt: time.Now().Add(ttl),
	}
	b.itemsMap[id] = item
	b.queue.Enqueue(item)
}

func (b *Bucket) Remove(id uint32) {
	b.rw.Lock()
	item, ok := b.itemsMap[id]
	if ok {
		delete(b.itemsMap, id)
		b.queue.Remove(item)
	}
	b.rw.Unlock()
}

func (b *Bucket) Peek() (*Item[uint32, bool], bool) {
	b.rw.RLock()
	head := b.queue[0]
	b.rw.RUnlock()
	return head, head == nil
}

func (b *Bucket) OnClose() {
	if b.timer != nil {
		b.timer.Stop()
	}
}

func (b *Bucket) check() {
	b.schedule()
	now := time.Now()
	ids := make([]uint32, 0, 1<<3)
	var num uint32
	b.rw.Lock()
	for {
		if b.queue.IsEmpty() {
			break
		}
		front := b.queue[0]
		if front.expiresAt.After(now) {
			break
		}
		delete(b.itemsMap, front.key)
		b.queue.Dequeue()
		ids = append(ids, front.key)
		num++
	}
	b.rw.Unlock()
	if num > 0 {
		b.callback(ids)
	}
}

func (b *Bucket) schedule() {
	if b.mgr.State() == TTypeRunning {
		b.timer = time.AfterFunc(time.Second, func() {
			b.check()
		})
	}
}
