package seqflow

import (
	"sync/atomic"
	"unsafe"
)

// sequence 是缓存行对齐的原子 int64，用于防止 false sharing。
// 不导出，防止用户创建未对齐的实例。通过 newSequence() 或 newSequences() 分配。
type sequence struct {
	_ [CacheLineBytes - unsafe.Sizeof(atomic.Int64{})]byte
	atomic.Int64
}

// defaultSequenceValue 是序列的默认初始值
const defaultSequenceValue = -1

// newSequence 分配一个缓存行对齐的序列，初始值为 -1
func newSequence() (this *sequence) {
	for this = new(sequence); uintptr(unsafe.Pointer(this))%CacheLineBytes != 0; this = new(sequence) {
	}
	this.Store(defaultSequenceValue)
	return this
}

// newSequences 分配一组连续的、缓存行对齐的序列
func newSequences(count int) []*sequence {
	var contiguous []sequence
	for contiguous = make([]sequence, count); uintptr(unsafe.Pointer(&contiguous[0]))%CacheLineBytes != 0; contiguous = make([]sequence, count) {
	}

	result := make([]*sequence, count)
	for i := range result {
		result[i] = &contiguous[i]
		result[i].Store(defaultSequenceValue)
	}
	return result
}
