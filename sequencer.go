package seqflow

// sentinel 错误码，内部使用，避免热路径上返回 error 元组
const (
	errReservationSize     = int64(-1) // 无效预留大小
	errCapacityUnavailable = int64(-2) // 容量不足
)

// singleSequencer 是单写者序列器，不支持并发写入。
// 热路径返回 int64 sentinel，避免 (int64, error) 元组开销。
type singleSequencer struct {
	_                      [4]byte         // 对齐填充
	capacity               uint32          // 环形缓冲区容量
	reservedSequence       int64           // 已预留的最高序列号（非原子，单写者无需同步）
	cachedConsumerSequence int64           // 本地缓存的最慢消费者位置
	committedSequence      *sequence       // 原子发布点
	consumerBarrier        SequenceBarrier // 查询最慢消费者
	waiter                 WaitStrategy    // 慢路径等待策略
}

func newSingleSequencer(capacity uint32, committed *sequence, consumerBarrier SequenceBarrier, waiter WaitStrategy) *singleSequencer {
	return &singleSequencer{
		capacity:               capacity,
		reservedSequence:       defaultSequenceValue,
		cachedConsumerSequence: defaultSequenceValue,
		committedSequence:      committed,
		consumerBarrier:        consumerBarrier,
		waiter:                 waiter,
	}
}

// reserve 阻塞式预留，返回上界序列号。错误时返回 sentinel 负值。
// 热路径无 error 元组、无 metrics、无 Signal — 纯序列逻辑。
func (s *singleSequencer) reserve(count uint32) int64 {
	if count == 0 || count > s.capacity {
		return errReservationSize
	}

	previousReserved := s.reservedSequence
	s.reservedSequence += int64(count)
	minimumSequence := s.reservedSequence - int64(s.capacity)

	// 快路径：检查本地缓存，无原子操作
	if minimumSequence <= s.cachedConsumerSequence && s.cachedConsumerSequence <= previousReserved {
		return s.reservedSequence
	}

	// 慢路径：自旋查询消费者屏障
	for spin := int64(0); ; spin++ {
		s.cachedConsumerSequence = s.consumerBarrier.Load(0)
		if minimumSequence <= s.cachedConsumerSequence {
			break
		}
		s.waiter.Reserve(spin)
	}

	return s.reservedSequence
}

// tryReserve 非阻塞式尝试预留
func (s *singleSequencer) tryReserve(count uint32) int64 {
	if count == 0 || count > s.capacity {
		return errReservationSize
	}

	previousReserved := s.reservedSequence
	s.reservedSequence += int64(count)
	minimumSequence := s.reservedSequence - int64(s.capacity)

	// 快路径
	if minimumSequence <= s.cachedConsumerSequence && s.cachedConsumerSequence <= previousReserved {
		return s.reservedSequence
	}

	// 慢路径：单次检查，不自旋
	s.cachedConsumerSequence = s.consumerBarrier.Load(0)
	if minimumSequence > s.cachedConsumerSequence {
		s.reservedSequence -= int64(count) // 回滚
		return errCapacityUnavailable
	}

	return s.reservedSequence
}

// commit 提交写入。纯原子 Store，无 Signal、无 metrics。
func (s *singleSequencer) commit(_, upper int64) {
	s.committedSequence.Store(upper)
}
