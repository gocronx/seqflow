package seqflow

import (
	"runtime"
	"sync"
	"time"
)

// WaitStrategy 控制生产者和消费者的背压行为
type WaitStrategy interface {
	// Gate 当数据已提交但上游 Handler 组未完成时调用
	Gate(count int64)
	// Idle 当没有数据可处理时调用
	Idle(count int64)
	// Reserve 当环形缓冲区已满、生产者等待时调用
	Reserve(count int64)
	// Signal 在生产者 Commit 时调用，用于唤醒阻塞的消费者
	Signal()
}

// BusySpinStrategy 忙等策略，不让出 CPU。适用于极低延迟、独占 CPU 核心的场景。
type BusySpinStrategy struct{}

func NewBusySpinStrategy() *BusySpinStrategy { return &BusySpinStrategy{} }
func (s *BusySpinStrategy) Gate(int64)       {}
func (s *BusySpinStrategy) Idle(int64)       {}
func (s *BusySpinStrategy) Reserve(int64)    {}
func (s *BusySpinStrategy) Signal()          {}

// YieldingStrategy 让出处理器策略。适用于低延迟、共享 CPU 的场景。
type YieldingStrategy struct{}

func NewYieldingStrategy() *YieldingStrategy { return &YieldingStrategy{} }
func (s *YieldingStrategy) Gate(int64)       { runtime.Gosched() }
func (s *YieldingStrategy) Idle(int64)       { runtime.Gosched() }
func (s *YieldingStrategy) Reserve(int64)    { runtime.Gosched() }
func (s *YieldingStrategy) Signal()          {}

// SleepingStrategy 默认策略。Gate 使用 Gosched（工作即将到来），Idle/Reserve 使用 Sleep。
type SleepingStrategy struct{}

func NewSleepingStrategy() *SleepingStrategy { return &SleepingStrategy{} }
func (s *SleepingStrategy) Gate(int64)       { runtime.Gosched() }
func (s *SleepingStrategy) Idle(int64)       { time.Sleep(500 * time.Nanosecond) }
func (s *SleepingStrategy) Reserve(int64)    { time.Sleep(time.Nanosecond) }
func (s *SleepingStrategy) Signal()          {}

// BlockingStrategy 阻塞策略，使用 sync.Cond 最小化 CPU 占用，延迟较高。
type BlockingStrategy struct {
	mu   sync.Mutex
	cond *sync.Cond
}

func NewBlockingStrategy() *BlockingStrategy {
	s := &BlockingStrategy{}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *BlockingStrategy) Gate(int64) {
	s.mu.Lock()
	s.cond.Wait()
	s.mu.Unlock()
}

func (s *BlockingStrategy) Idle(int64) {
	s.mu.Lock()
	s.cond.Wait()
	s.mu.Unlock()
}

func (s *BlockingStrategy) Reserve(int64) {
	s.mu.Lock()
	s.cond.Wait()
	s.mu.Unlock()
}

func (s *BlockingStrategy) Signal() {
	s.cond.Broadcast()
}
