package seqflow

import "testing"

func TestAtomicBarrier_Load(t *testing.T) {
	seq := newSequence()
	seq.Store(42)
	barrier := newAtomicBarrier(seq)
	if got := barrier.Load(0); got != 42 {
		t.Errorf("atomicBarrier.Load() = %d, want 42", got)
	}
}

func TestCompositeBarrier_ReturnsMinimum(t *testing.T) {
	seqs := newSequences(3)
	seqs[0].Store(10)
	seqs[1].Store(5)
	seqs[2].Store(8)
	barrier := newCompositeBarrier(seqs...)
	if got := barrier.Load(0); got != 5 {
		t.Errorf("compositeBarrier.Load() = %d, want 5", got)
	}
}

func TestCompositeBarrier_SingleOptimization(t *testing.T) {
	seq := newSequence()
	seq.Store(99)
	barrier := newCompositeBarrier(seq)
	if _, ok := barrier.(atomicBarrier); !ok {
		t.Errorf("单序列 compositeBarrier 应返回 atomicBarrier，实际为 %T", barrier)
	}
	if got := barrier.Load(0); got != 99 {
		t.Errorf("barrier.Load() = %d, want 99", got)
	}
}

func TestCompositeBarrier_Empty(t *testing.T) {
	barrier := newCompositeBarrier()
	got := barrier.Load(0)
	if got != 1<<63-1 {
		t.Errorf("空 compositeBarrier.Load() = %d, want MaxInt64", got)
	}
}

func TestCompositeBarrier_UpdatesPropagated(t *testing.T) {
	seqs := newSequences(2)
	seqs[0].Store(10)
	seqs[1].Store(20)
	barrier := newCompositeBarrier(seqs...)

	if got := barrier.Load(0); got != 10 {
		t.Errorf("初始 Load() = %d, want 10", got)
	}

	seqs[0].Store(25)
	if got := barrier.Load(0); got != 20 {
		t.Errorf("更新后 Load() = %d, want 20", got)
	}
}
