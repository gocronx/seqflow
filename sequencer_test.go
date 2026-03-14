package seqflow

import "testing"

func TestSingleSequencer_Reserve_Basic(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSingleSequencer(1024, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	upper := s.reserve(1)
	if upper != 0 {
		t.Errorf("首次 reserve(1) = %d, 期望 0", upper)
	}
}

func TestSingleSequencer_Reserve_Batch(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSingleSequencer(1024, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	upper := s.reserve(16)
	if upper != 15 {
		t.Errorf("reserve(16) = %d, 期望 15", upper)
	}
}

func TestSingleSequencer_Reserve_InvalidSize(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	s := newSingleSequencer(1024, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	if got := s.reserve(0); got != errReservationSize {
		t.Errorf("reserve(0) = %d, 期望 sentinel %d", got, errReservationSize)
	}
	if got := s.reserve(2048); got != errReservationSize {
		t.Errorf("reserve(2048) = %d, 期望 sentinel %d", got, errReservationSize)
	}
}

func TestSingleSequencer_TryReserve_Success(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSingleSequencer(1024, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	upper := s.tryReserve(1)
	if upper != 0 {
		t.Errorf("tryReserve(1) = %d, 期望 0", upper)
	}
}

func TestSingleSequencer_TryReserve_Unavailable(t *testing.T) {
	committed := newSequence()
	consumer := newSequence() // 消费者位于 -1
	s := newSingleSequencer(4, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	// 填满缓冲区
	for i := uint32(0); i < 4; i++ {
		s.reserve(1)
	}

	if got := s.tryReserve(1); got != errCapacityUnavailable {
		t.Errorf("缓冲区满时 tryReserve = %d, 期望 sentinel %d", got, errCapacityUnavailable)
	}
}

func TestSingleSequencer_Commit(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	consumer.Store(1024)
	s := newSingleSequencer(1024, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	upper := s.reserve(1)
	s.commit(upper, upper)

	if got := committed.Load(); got != upper {
		t.Errorf("已提交序列 = %d, 期望 %d", got, upper)
	}
}

func TestSingleSequencer_Reserve_WrapAround(t *testing.T) {
	committed := newSequence()
	consumer := newSequence()
	s := newSingleSequencer(4, committed, newAtomicBarrier(consumer), NewSleepingStrategy())

	// 填满缓冲区并推进消费者
	for i := uint32(0); i < 4; i++ {
		upper := s.reserve(1)
		s.commit(upper, upper)
		consumer.Store(upper)
	}

	// 回绕后应成功
	upper := s.reserve(1)
	if upper < 0 {
		t.Fatalf("回绕后 reserve 返回 sentinel: %d", upper)
	}
	if upper != 4 {
		t.Errorf("回绕后 reserve = %d, 期望 4", upper)
	}
}
