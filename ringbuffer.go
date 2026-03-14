package seqflow

import "fmt"

// RingBuffer 是泛型的预分配环形缓冲区，容量必须是 2 的幂。
// Get 返回指向缓冲区槽位的指针（零拷贝），该指针仅在生产者回绕覆写该槽位前有效。
type RingBuffer[T any] struct {
	buffer   []T
	capacity uint32
	mask     uint32
}

// NewRingBuffer 创建指定容量的环形缓冲区（必须是 2 的幂）
func NewRingBuffer[T any](capacity uint32) (*RingBuffer[T], error) {
	if capacity == 0 || capacity&(capacity-1) != 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidCapacity, capacity)
	}
	return &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
		mask:     capacity - 1,
	}, nil
}

// Get 返回指定序列位置的元素指针
func (rb *RingBuffer[T]) Get(seq int64) *T {
	return &rb.buffer[seq&int64(rb.mask)]
}

// Set 写入值到指定序列位置
func (rb *RingBuffer[T]) Set(seq int64, value T) {
	rb.buffer[seq&int64(rb.mask)] = value
}

// Capacity 返回环形缓冲区容量
func (rb *RingBuffer[T]) Capacity() uint32 {
	return rb.capacity
}
