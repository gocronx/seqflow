package seqflow

import "testing"

func TestBusySpinStrategy_DoesNotPanic(t *testing.T) {
	s := NewBusySpinStrategy()
	s.Gate(1)
	s.Idle(1)
	s.Reserve(1)
	s.Signal()
}

func TestYieldingStrategy_DoesNotPanic(t *testing.T) {
	s := NewYieldingStrategy()
	s.Gate(1)
	s.Idle(1)
	s.Reserve(1)
	s.Signal()
}

func TestSleepingStrategy_DoesNotPanic(t *testing.T) {
	s := NewSleepingStrategy()
	s.Gate(1)
	s.Idle(1)
	s.Reserve(1)
	s.Signal()
}

func TestBlockingStrategy_SignalWakesWaiter(t *testing.T) {
	s := NewBlockingStrategy()
	// Signal 不应在无等待者时 panic
	s.Signal()

	done := make(chan struct{})
	go func() {
		s.Idle(1)
		close(done)
	}()

	// 持续 Signal 直到 goroutine 被唤醒
	for {
		select {
		case <-done:
			return
		default:
			s.Signal()
		}
	}
}

func TestWaitStrategy_InterfaceCompliance(t *testing.T) {
	var _ WaitStrategy = NewBusySpinStrategy()
	var _ WaitStrategy = NewYieldingStrategy()
	var _ WaitStrategy = NewSleepingStrategy()
	var _ WaitStrategy = NewBlockingStrategy()
}
