package rpc

/*
   @Author: orbit-w
   @File: expiration_queue
   @2023 12月 周日 19:40
*/

import (
	"container/heap"
	"time"
)

type Item[K comparable, V any] struct {
	key       K
	value     V
	index     int
	ttl       time.Duration
	expiresAt time.Time
}

func (i Item[K, V]) Equal(dst *Item[K, V]) bool {
	return i.key == dst.key
}

type expirationQueue[K comparable, V any] []*Item[K, V]

func newExpirationQueue[K comparable, V any]() expirationQueue[K, V] {
	q := make(expirationQueue[K, V], 0)
	heap.Init(&q)
	return q
}

// IsEmpty checks if the queue is empty.
func (q expirationQueue[K, V]) IsEmpty() bool {
	return q.Len() == 0
}

// Update updates an existing item's value and position in the queue.
func (q *expirationQueue[K, V]) Update(item *Item[K, V]) {
	heap.Fix(q, item.index)
}

// Enqueue pushes a new item into the queue and updates the order of its
// elements.
func (q *expirationQueue[K, V]) Enqueue(item *Item[K, V]) {
	heap.Push(q, item)
}

// Remove removes an item from the queue and updates the order of its
// elements.
func (q *expirationQueue[K, V]) Remove(item *Item[K, V]) {
	heap.Remove(q, item.index)
}

func (q *expirationQueue[K, V]) Dequeue() *Item[K, V] {
	v := heap.Pop(q)
	if v == nil {
		return nil
	}
	return v.(*Item[K, V])
}

func (q *expirationQueue[K, V]) Peek() *Item[K, V] {
	if q.Len() > 0 {
		return (*q)[0]
	} else {
		return nil
	}
}

// Len returns the total number of items in the queue.
func (q expirationQueue[K, V]) Len() int {
	return len(q)
}

// Less checks if the item at the i position expires sooner than
// the one at the j position.
func (q expirationQueue[K, V]) Less(i, j int) bool {
	item1, item2 := q[i], q[j]
	if item1.expiresAt.IsZero() {
		return false
	}

	if item2.expiresAt.IsZero() {
		return true
	}

	return item1.expiresAt.Before(item2.expiresAt)
}

// Swap switches the places of two queue items.
func (q expirationQueue[K, V]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

// Push appends a new item to the item slice.
func (q *expirationQueue[K, V]) Push(x interface{}) {
	item := x.(*Item[K, V])
	item.index = len(*q)
	*q = append(*q, item)
}

// Pop removes and returns the last item.
func (q *expirationQueue[K, V]) Pop() interface{} {
	old := *q
	i := len(old) - 1
	item := old[i]
	item.index = -1
	old[i] = nil // avoid memory leak
	*q = old[:i]

	return item
}
