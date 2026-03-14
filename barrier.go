package seqflow

// SequenceBarrier 抽象了从一个或多个序列读取已提交/已处理位置的操作。
type SequenceBarrier interface {
	// Load 返回从给定下界开始可安全读取的最高序列号
	Load(lower int64) int64
}

// atomicBarrier 包装单个序列作为屏障
type atomicBarrier struct {
	seq *sequence
}

func newAtomicBarrier(seq *sequence) atomicBarrier {
	return atomicBarrier{seq: seq}
}

func (b atomicBarrier) Load(_ int64) int64 {
	return b.seq.Load()
}

// compositeBarrier 返回多个序列中的最小值（最慢消费者的位置）
type compositeBarrier []*sequence

func newCompositeBarrier(seqs ...*sequence) SequenceBarrier {
	switch len(seqs) {
	case 0:
		return compositeBarrier{}
	case 1:
		return newAtomicBarrier(seqs[0])
	default:
		return compositeBarrier(seqs)
	}
}

func (b compositeBarrier) Load(_ int64) int64 {
	minimum := int64(1<<63 - 1)
	for _, seq := range b {
		if v := seq.Load(); v < minimum {
			minimum = v
		}
	}
	return minimum
}
