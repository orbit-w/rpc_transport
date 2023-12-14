package ringbuffer_v2

type RingBuffer struct {
	data []interface{}
	cap  int
	head int
	tail int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data: make([]interface{}, capacity),
		cap:  capacity,
	}
}

func (rb *RingBuffer) IsEmpty() bool {
	return rb.head == rb.tail
}

func (rb *RingBuffer) Size() int {
	if rb.head >= rb.tail {
		return rb.head - rb.tail
	} else {
		return rb.cap + rb.head - rb.tail
	}
}

func (rb *RingBuffer) IsFull() bool {
	return ((rb.head + 1) % rb.cap) == (rb.tail % rb.cap)
}

// Peek returns the first element in ring buffer without removing it.
// If ring buffer is empty, returns nil.
func (rb *RingBuffer) Peek() interface{} {

	if rb.IsEmpty() {
		return nil
	}
	return rb.data[rb.tail]
}

// Put adds an element to the end of the ring buffer. If it's full, double its size.
func (rb *RingBuffer) Put(val interface{}) {

	if !rb.IsFull() {
		rb.data[rb.head] = val
		rb.head = (rb.head + 1) % rb.cap
	} else {
		oldCap := rb.cap
		newData := make([]interface{}, rb.cap*2)

		for i := 0; i < oldCap; i++ {
			newData[i] = rb.data[(i+rb.tail)%oldCap]
		}

		// replace with new slice.
		rb.data = newData

		// reset head and tail pointer after resizing.
		rb.tail = 0
		rb.head = oldCap

		// update cap after resizing.
		rb.cap *= 2

		rb.Put(val)
	}

}

func (rb *RingBuffer) Pop() interface{} {
	if rb.IsEmpty() {
		return nil
	}

	head := rb.data[rb.tail]
	rb.data[rb.tail] = nil
	rb.tail = (rb.tail + 1) % rb.cap
	return head
}
