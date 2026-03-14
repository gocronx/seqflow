package seqflow

import (
	"sync"
	"testing"
)

func TestSharedSequencer_Reserve_Basic(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSharedSequencer(1024, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	upper := s.reserve(1)
	if upper != 0 {
		t.Errorf("首次 reserve(1) = %d, 期望 0", upper)
	}
}

func TestSharedSequencer_Reserve_InvalidSize(t *testing.T) {
	reserved := newSequence()
	s := newSharedSequencer(1024, reserved, NewSleepingStrategy())

	if got := s.reserve(0); got != errReservationSize {
		t.Errorf("reserve(0) = %d, 期望 sentinel", got)
	}
	if got := s.reserve(2048); got != errReservationSize {
		t.Errorf("reserve(2048) = %d, 期望 sentinel", got)
	}
}

func TestSharedSequencer_TryReserve_Success(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSharedSequencer(1024, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	upper := s.tryReserve(1)
	if upper != 0 {
		t.Errorf("首次 tryReserve(1) = %d, 期望 0", upper)
	}
}

func TestSharedSequencer_TryReserve_Unavailable(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence() // 消费者位于 -1
	s := newSharedSequencer(4, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	// 填满缓冲区
	for i := uint32(0); i < 4; i++ {
		s.reserve(1)
	}

	if got := s.tryReserve(1); got != errCapacityUnavailable {
		t.Errorf("缓冲区满时 tryReserve = %d, 期望 sentinel", got)
	}
}

func TestSharedSequencer_Commit_And_Load(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSharedSequencer(8, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	// 顺序预留并提交 3 个槽位
	for i := 0; i < 3; i++ {
		upper := s.reserve(1)
		s.commit(upper, upper)
	}

	if got := s.Load(0); got != 2 {
		t.Errorf("Load(0) = %d, 期望 2", got)
	}
}

func TestSharedSequencer_Commit_OutOfOrder(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSharedSequencer(8, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	upper0 := s.reserve(1) // 序列 0
	upper1 := s.reserve(1) // 序列 1
	upper2 := s.reserve(1) // 序列 2

	// 乱序提交：先 2，再 0，最后 1
	s.commit(upper2, upper2)
	if got := s.Load(0); got != -1 {
		t.Errorf("仅提交序列 2 后 Load = %d, 期望 -1", got)
	}

	s.commit(upper0, upper0)
	if got := s.Load(0); got != 0 {
		t.Errorf("提交序列 0 和 2 后 Load = %d, 期望 0", got)
	}

	s.commit(upper1, upper1)
	if got := s.Load(0); got != 2 {
		t.Errorf("全部提交后 Load = %d, 期望 2", got)
	}
}

func TestSharedSequencer_ConcurrentReserve(t *testing.T) {
	reserved := newSequence()
	consumer := newSequence()
	consumer.Store(100000)
	s := newSharedSequencer(1024, reserved, NewSleepingStrategy())
	s.consumerBarrier = newAtomicBarrier(consumer)

	const goroutines = 4
	const perGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				upper := s.reserve(1)
				if upper < 0 {
					t.Errorf("并发 reserve 返回 sentinel: %d", upper)
					return
				}
				s.commit(upper, upper)
			}
		}()
	}
	wg.Wait()

	expected := int64(goroutines*perGoroutine - 1)
	if got := reserved.Load(); got != expected {
		t.Errorf("预留序列 = %d, 期望 %d", got, expected)
	}
}
