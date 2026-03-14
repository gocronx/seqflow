package seqflow

import (
	"testing"
	"unsafe"
)

// TestSequenceDefaultValue 验证新序列的默认值为 -1
func TestSequenceDefaultValue(t *testing.T) {
	seq := newSequence()
	if got := seq.Load(); got != defaultSequenceValue {
		t.Errorf("期望默认值 %d，实际得到 %d", defaultSequenceValue, got)
	}
}

// TestSequenceCacheLineAlignment 验证序列地址按缓存行对齐
func TestSequenceCacheLineAlignment(t *testing.T) {
	seq := newSequence()
	addr := uintptr(unsafe.Pointer(seq))
	if addr%CacheLineBytes != 0 {
		t.Errorf("序列地址 %x 未按 %d 字节对齐", addr, CacheLineBytes)
	}
}

// TestSequenceStoreLoad 验证存储和加载操作
func TestSequenceStoreLoad(t *testing.T) {
	seq := newSequence()
	seq.Store(42)
	if got := seq.Load(); got != 42 {
		t.Errorf("期望值 42，实际得到 %d", got)
	}

	seq.Store(1000000)
	if got := seq.Load(); got != 1000000 {
		t.Errorf("期望值 1000000，实际得到 %d", got)
	}
}

// TestNewSequencesCount 验证 newSequences 返回正确数量的序列
func TestNewSequencesCount(t *testing.T) {
	count := 5
	seqs := newSequences(count)
	if len(seqs) != count {
		t.Errorf("期望 %d 个序列，实际得到 %d 个", count, len(seqs))
	}
}

// TestNewSequencesAlignment 验证批量分配的序列均按缓存行对齐
func TestNewSequencesAlignment(t *testing.T) {
	seqs := newSequences(4)
	for i, seq := range seqs {
		addr := uintptr(unsafe.Pointer(seq))
		if addr%CacheLineBytes != 0 {
			t.Errorf("序列[%d] 地址 %x 未按 %d 字节对齐", i, addr, CacheLineBytes)
		}
	}
}

// TestNewSequencesContiguity 验证批量分配的序列在内存中连续
func TestNewSequencesContiguity(t *testing.T) {
	seqs := newSequences(4)
	seqSize := unsafe.Sizeof(sequence{})
	for i := 1; i < len(seqs); i++ {
		prev := uintptr(unsafe.Pointer(seqs[i-1]))
		curr := uintptr(unsafe.Pointer(seqs[i]))
		if curr-prev != seqSize {
			t.Errorf("序列[%d] 和 序列[%d] 不连续：间距 %d，期望 %d", i-1, i, curr-prev, seqSize)
		}
	}
}

// TestNewSequencesDefaultValues 验证批量分配的序列默认值均为 -1
func TestNewSequencesDefaultValues(t *testing.T) {
	seqs := newSequences(3)
	for i, seq := range seqs {
		if got := seq.Load(); got != defaultSequenceValue {
			t.Errorf("序列[%d] 期望默认值 %d，实际得到 %d", i, defaultSequenceValue, got)
		}
	}
}
