package seqflow

import (
	"sync"
	"sync/atomic"
)

const (
	stateRunning = int64(0)
	stateClosed  = int64(1)
)

// listenCloser 处理环形缓冲区中的事件
type listenCloser interface {
	Listen()
	Close() error
}

// listener 是每个 handler 的消费者循环，非 goroutine 安全。
// 三状态主循环：有数据→处理 / 被 gate→退避 / 空闲→等待
type listener struct {
	name             string          // handler 名称，用于 metrics
	running          atomic.Int64    // 0=运行中, 1=已关闭
	handledSequence  *sequence       // 该消费者的处理进度
	committedBarrier SequenceBarrier // 生产者提交进度
	upstreamBarrier  SequenceBarrier // 上游依赖的进度
	waiter           WaitStrategy    // 背压策略
	handler          Handler         // 用户回调
	metrics          Metrics         // 可选指标
}

func newListener(name string, handledSeq *sequence, committedBarrier, upstreamBarrier SequenceBarrier, waiter WaitStrategy, handler Handler, metrics Metrics) *listener {
	return &listener{
		name:             name,
		handledSequence:  handledSeq,
		committedBarrier: committedBarrier,
		upstreamBarrier:  upstreamBarrier,
		waiter:           waiter,
		handler:          handler,
		metrics:          metrics,
	}
}

// Listen 阻塞当前 goroutine，持续处理事件直到被关闭
func (l *listener) Listen() {
	var gatedCount, idlingCount int64
	handledSequence := l.handledSequence.Load()

	for {
		lowerSequence := handledSequence + 1
		upperSequence := l.upstreamBarrier.Load(lowerSequence)

		if lowerSequence <= upperSequence {
			// 有数据可处理
			l.handler.Handle(lowerSequence, upperSequence)
			l.handledSequence.Store(upperSequence)
			handledSequence = upperSequence
			if l.metrics != nil {
				l.metrics.HandleCount(l.name, 1)
				l.metrics.HandleEvents(l.name, upperSequence-lowerSequence+1)
			}
			gatedCount = 0
			idlingCount = 0
		} else if upperSequence = l.committedBarrier.Load(lowerSequence); lowerSequence <= upperSequence {
			// 数据已提交但上游未完成，被 gate
			gatedCount++
			idlingCount = 0
			if l.metrics != nil {
				l.metrics.GateCount(l.name, 1)
			}
			l.waiter.Gate(gatedCount)
		} else if l.running.Load() == stateRunning {
			// 空闲，无数据
			idlingCount++
			gatedCount = 0
			if l.metrics != nil {
				l.metrics.IdleCount(l.name, 1)
			}
			l.waiter.Idle(idlingCount)
		} else {
			// 已关闭，退出循环
			break
		}
	}
}

// Close 停止消费者循环
func (l *listener) Close() error {
	l.running.Store(stateClosed)
	return nil
}

// compositeListener 管理多个 listener 并发运行
type compositeListener []listenCloser

func newCompositeListener(listeners []listenCloser) listenCloser {
	if len(listeners) == 1 {
		return listeners[0]
	}
	return compositeListener(listeners)
}

// Listen 在独立 goroutine 上启动每个子 listener，等待全部完成
func (cl compositeListener) Listen() {
	var wg sync.WaitGroup
	wg.Add(len(cl))

	for _, item := range cl {
		go func(l listenCloser) {
			defer wg.Done()
			l.Listen()
		}(item)
	}

	wg.Wait()
}

// Close 关闭所有子 listener
func (cl compositeListener) Close() error {
	for _, item := range cl {
		_ = item.Close()
	}
	return nil
}
