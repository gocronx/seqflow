package seqflow

import (
	"sync"
	"testing"
	"time"
)

type recordingHandler struct {
	handled []int64
	mu      sync.Mutex
}

func (h *recordingHandler) Handle(lower, upper int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := lower; i <= upper; i++ {
		h.handled = append(h.handled, i)
	}
}

func TestListener_ProcessesEvents(t *testing.T) {
	committed := newSequence()
	handler := &recordingHandler{}
	handledSeq := newSequence()

	l := newListener("test", handledSeq, newAtomicBarrier(committed), newAtomicBarrier(committed), NewSleepingStrategy(), handler, nil)

	go l.Listen()

	// 发布 3 个事件
	committed.Store(2)

	// 等待消费者处理
	deadline := time.Now().Add(time.Second)
	for handledSeq.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	l.Close()

	if got := handledSeq.Load(); got != 2 {
		t.Errorf("handledSequence = %d, 期望 2", got)
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()
	if len(handler.handled) != 3 {
		t.Errorf("处理了 %d 个事件, 期望 3", len(handler.handled))
	}
}

func TestListener_Close(t *testing.T) {
	committed := newSequence()
	handler := &recordingHandler{}
	handledSeq := newSequence()

	l := newListener("test", handledSeq, newAtomicBarrier(committed), newAtomicBarrier(committed), NewSleepingStrategy(), handler, nil)

	done := make(chan struct{})
	go func() {
		l.Listen()
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	l.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close 后 Listen 未返回")
	}
}

// handlerFunc 将函数适配为 Handler 接口
type handlerFunc func(lower, upper int64)

func (f handlerFunc) Handle(lower, upper int64) { f(lower, upper) }
