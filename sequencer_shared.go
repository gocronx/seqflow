package seqflow

import (
	"math/bits"
	"sync/atomic"
)

// sharedSequencer 是多写者序列器，允许多个 goroutine 并发生产。
// 占两个缓存行：第一行用于热路径（reserve/commit/Load），第二行用于慢路径。
// 热路径返回 int64 sentinel，避免 (int64, error) 元组开销。
type sharedSequencer struct {
	// 缓存行 1 — 热路径
	reservedSequence       *sequence      // 原子 Add 抢占槽位
	cachedConsumerSequence *sequence      // 原子缓存的消费者位置
	committedSlots         []atomic.Int32 // 每槽位 round 追踪，支持乱序提交
	capacity               uint32         // 缓冲区容量（2 的幂）
	shift                  uint8          // log2(capacity)，用于计算 round
	_                      [19]byte       // 填充至 64 字节边界

	// 缓存行 2 — 慢路径
	consumerBarrier SequenceBarrier // 查询最慢消费者
	waiter          WaitStrategy    // 慢路径等待策略
	_               [32]byte        // 尾部填充
}

func newSharedSequencer(capacity uint32, reserved *sequence, waiter WaitStrategy) *sharedSequencer {
	slots := make([]atomic.Int32, capacity)
	for i := range slots {
		slots[i].Store(int32(defaultSequenceValue))
	}

	return &sharedSequencer{
		reservedSequence:       reserved,
		cachedConsumerSequence: newSequence(),
		committedSlots:         slots,
		capacity:               capacity,
		shift:                  uint8(bits.TrailingZeros32(capacity)),
		waiter:                 waiter,
	}
}

// reserve 阻塞式预留，使用 atomic.Add 抢占（无 CAS 争用）。
// 注意：atomic.Add 可能在高争用时导致临时过度预留，自旋等待消费者推进后自然恢复。
func (s *sharedSequencer) reserve(count uint32) int64 {
	if count == 0 || count > s.capacity {
		return errReservationSize
	}

	slots := int64(count)
	reservedSequence := s.reservedSequence.Add(slots)
	previousReserved := reservedSequence - slots
	minimumSequence := reservedSequence - int64(s.capacity)
	consumerSequence := s.cachedConsumerSequence.Load()

	// 快路径
	if minimumSequence <= consumerSequence && consumerSequence <= previousReserved {
		return reservedSequence
	}

	// 慢路径：自旋查询消费者屏障
	for spin := int64(0); ; spin++ {
		consumerSequence = s.consumerBarrier.Load(0)
		if minimumSequence <= consumerSequence {
			break
		}
		s.waiter.Reserve(spin)
	}

	s.cachedConsumerSequence.Store(consumerSequence)
	return reservedSequence
}

// tryReserve 单次 CAS 尝试，失败立即返回
func (s *sharedSequencer) tryReserve(count uint32) int64 {
	if count == 0 || count > s.capacity {
		return errReservationSize
	}

	slots := int64(count)
	previousReserved := s.reservedSequence.Load()
	if !s.hasAvailableCapacity(previousReserved, slots) {
		return errCapacityUnavailable
	}

	if !s.reservedSequence.CompareAndSwap(previousReserved, previousReserved+slots) {
		return errCapacityUnavailable
	}

	return previousReserved + slots
}

// hasAvailableCapacity 检查是否有足够容量
func (s *sharedSequencer) hasAvailableCapacity(previousReserved, count int64) bool {
	reservedSequence := previousReserved + count
	minimumSequence := reservedSequence - int64(s.capacity)
	consumerSequence := s.cachedConsumerSequence.Load()

	if minimumSequence <= consumerSequence && consumerSequence <= previousReserved {
		return true
	}

	consumerSequence = s.consumerBarrier.Load(0)
	s.cachedConsumerSequence.Store(consumerSequence)
	return minimumSequence <= consumerSequence
}

// commit 提交写入，每个槽位存储 round 值以支持乱序提交。纯数据操作，无 Signal/metrics。
func (s *sharedSequencer) commit(lower, upper int64) {
	mask := int64(s.capacity) - 1
	for ; lower <= upper; lower++ {
		s.committedSlots[lower&mask].Store(int32(lower >> s.shift))
	}
}

// Load 返回从 lower 开始连续已提交的最高序列号。
// sharedSequencer 同时充当第一个消费者组的 committedBarrier。
func (s *sharedSequencer) Load(lower int64) int64 {
	upper := s.reservedSequence.Load()
	mask := int64(s.capacity) - 1

	for ; lower <= upper; lower++ {
		if s.committedSlots[lower&mask].Load() != int32(lower>>s.shift) {
			return lower - 1
		}
	}

	return upper
}
