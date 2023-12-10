package bounded

/*
   @Author: orbit-w
   @File: bounded
   @2023 12月 周日 11:03
*/

import (
	rbqueue "github.com/Workiva/go-datastructures/queue"
)

type BoundedQueue struct {
	userMailbox *rbqueue.RingBuffer
	dropping    bool
}

func New(size uint64) *BoundedQueue {
	return &BoundedQueue{
		userMailbox: rbqueue.NewRingBuffer(size),
	}
}

func (q *BoundedQueue) Push(m interface{}) {
	if q.dropping {
		if q.userMailbox.Len() > 0 && q.userMailbox.Cap()-1 == q.userMailbox.Len() {
			_, _ = q.userMailbox.Get()
		}
	}

	_ = q.userMailbox.Put(m)
}

func (q *BoundedQueue) Pop() interface{} {
	if q.userMailbox.Len() > 0 {
		m, _ := q.userMailbox.Get()
		return m
	}
	return nil
}
