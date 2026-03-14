package seqflow

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// Disruptor 是序列驱动流的顶层容器。
//
// 核心优化：
//  1. 预计算剩余容量：快路径仅 1 次比较
//  2. 零接口分发：单写者字段直接嵌入结构体
//  3. 零原子读：关闭时毒化 remainingCapacity 迫使进入慢路径
type Disruptor[T any] struct {
	// ===== 单写者热路径字段（直接嵌入，零间接引用）=====
	reservedSequence       int64          // 已预留的最高序列号
	remainingCapacity      int64          // 预计算剩余容量，快路径单次比较
	committedStore         *atomic.Int64  // 预解引用的 atomic 指针（消除 Commit 的指针追逐）
	cachedConsumerSequence int64          // 本地缓存的最慢消费者位置
	capacity               uint32         // 环形缓冲区容量

	// ===== 多写者路径（nil 表示单写者模式）=====
	shared *sharedSequencer

	// ===== 慢路径字段 =====
	singleConsumerBarrier SequenceBarrier // 单写者的消费者屏障
	singleWaiter          WaitStrategy    // 单写者的等待策略
	committedSeq          *sequence       // listener barrier 读取用
	signal                func()          // 非 nil 时在 Commit 后调用（仅 BlockingStrategy）

	// ===== 通用字段 =====
	ringBuffer      *RingBuffer[T]
	listener        listenCloser
	closed          atomic.Int64
	terminalBarrier SequenceBarrier
	metrics         Metrics
}

// Option 配置 Disruptor
type Option func(*config)

type config struct {
	capacity     uint32
	writerCount  uint8
	waitStrategy WaitStrategy
	metrics      Metrics
	handlers     []handlerNode
}

// WithCapacity 设置环形缓冲区容量（必须是 2 的幂）
func WithCapacity(n uint32) Option {
	return func(c *config) { c.capacity = n }
}

// WithWriterCount 设置并发生产者数量
func WithWriterCount(n uint8) Option {
	return func(c *config) { c.writerCount = n }
}

// WithWaitStrategy 设置背压策略
func WithWaitStrategy(ws WaitStrategy) Option {
	return func(c *config) { c.waitStrategy = ws }
}

// WithMetrics 设置指标收集器
func WithMetrics(m Metrics) Option {
	return func(c *config) { c.metrics = m }
}

// HandlerOption 配置 handler 注册
type HandlerOption func(*handlerNode)

// DependsOn 声明对其他命名 handler 的依赖
func DependsOn(names ...string) HandlerOption {
	return func(n *handlerNode) { n.dependsOn = append(n.dependsOn, names...) }
}

// WithHandler 注册一个命名的 handler 及其可选依赖
func WithHandler(name string, h Handler, opts ...HandlerOption) Option {
	return func(c *config) {
		node := handlerNode{name: name, handler: h}
		for _, opt := range opts {
			opt(&node)
		}
		c.handlers = append(c.handlers, node)
	}
}

// New 创建一个 Disruptor 实例
func New[T any](opts ...Option) (*Disruptor[T], error) {
	cfg := &config{
		capacity:     1024,
		writerCount:  1,
		waitStrategy: NewSleepingStrategy(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.capacity == 0 || cfg.capacity&(cfg.capacity-1) != 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidCapacity, cfg.capacity)
	}
	if len(cfg.handlers) == 0 {
		return nil, ErrNoHandlers
	}

	topology, err := buildDAG(cfg.handlers)
	if err != nil {
		return nil, err
	}

	rb, _ := NewRingBuffer[T](cfg.capacity)
	committedSeq := newSequence()

	var shared *sharedSequencer
	var committedBarrier SequenceBarrier

	if cfg.writerCount <= 1 {
		committedBarrier = newAtomicBarrier(committedSeq)
	} else {
		shared = newSharedSequencer(cfg.capacity, committedSeq, cfg.waitStrategy)
		committedBarrier = sharedBarrierAdapter{shared}
	}

	// 为所有 handler 分配序列
	allSequences := newSequences(len(topology.order))
	seqMap := make(map[string]*sequence, len(topology.order))
	for i, name := range topology.order {
		seqMap[name] = allSequences[i]
	}

	// 按拓扑顺序构建 listener
	listenerMap := make(map[string]listenCloser, len(topology.order))
	for _, name := range topology.order {
		var node handlerNode
		for _, n := range topology.nodes {
			if n.name == name {
				node = n
				break
			}
		}

		var upstreamBarrier SequenceBarrier
		if len(node.dependsOn) == 0 {
			upstreamBarrier = committedBarrier
		} else {
			deps := make([]*sequence, len(node.dependsOn))
			for i, dep := range node.dependsOn {
				deps[i] = seqMap[dep]
			}
			upstreamBarrier = newCompositeBarrier(deps...)
		}

		listenerMap[name] = newListener(name, seqMap[name], committedBarrier, upstreamBarrier, cfg.waitStrategy, node.handler, cfg.metrics)
	}

	// 终端屏障
	terminalSeqs := make([]*sequence, len(topology.terminals))
	for i, name := range topology.terminals {
		terminalSeqs[i] = seqMap[name]
	}
	terminalBarrier := newCompositeBarrier(terminalSeqs...)

	if cfg.writerCount <= 1 && shared == nil {
		// 单写者不需要单独的 singleSequencer 对象，字段直接嵌入 Disruptor
	} else if shared != nil {
		shared.consumerBarrier = terminalBarrier
	}

	// 收集所有 listener
	allListeners := make([]listenCloser, len(topology.order))
	for i, name := range topology.order {
		allListeners[i] = listenerMap[name]
	}

	var signalFn func()
	if bs, ok := cfg.waitStrategy.(*BlockingStrategy); ok {
		signalFn = bs.Signal
	}

	return &Disruptor[T]{
		// 单写者热路径
		reservedSequence:       defaultSequenceValue,
		remainingCapacity:      int64(cfg.capacity),
		committedStore:         &committedSeq.Int64, // 预解引用：Commit 直接写此地址
		cachedConsumerSequence: defaultSequenceValue,
		capacity:               cfg.capacity,
		// 多写者
		shared: shared,
		// 慢路径
		singleConsumerBarrier: terminalBarrier,
		singleWaiter:          cfg.waitStrategy,
		committedSeq:          committedSeq,
		signal:                signalFn,
		// 通用
		ringBuffer:      rb,
		listener:        newCompositeListener(allListeners),
		terminalBarrier: terminalBarrier,
		metrics:         cfg.metrics,
	}, nil
}

// sharedBarrierAdapter 将 sharedSequencer 包装为 SequenceBarrier
type sharedBarrierAdapter struct{ s *sharedSequencer }

func (a sharedBarrierAdapter) Load(lower int64) int64 { return a.s.Load(lower) }

// RingBuffer 返回底层环形缓冲区
func (d *Disruptor[T]) RingBuffer() *RingBuffer[T] { return d.ringBuffer }

// Reserve 在环形缓冲区中预留槽位。
//
// 单写者快路径（最常见场景）：
//   - 1 次 nil 检查（分支预测命中）
//   - 1 次比较（remainingCapacity）
//   - 2 次加减法
//   - 零原子操作，零接口分发，零 error 构造
func (d *Disruptor[T]) Reserve(count uint32) (int64, error) {
	if d.shared != nil {
		return d.reserveShared(count)
	}
	// 单写者快路径：预计算容量 + 直接字段访问
	slots := int64(count)
	if slots <= d.remainingCapacity {
		d.reservedSequence += slots
		d.remainingCapacity -= slots
		if d.metrics != nil {
			d.metrics.ReserveCount(1)
		}
		return d.reservedSequence, nil
	}
	return d.reserveSingleSlow(count)
}

// reserveShared 多写者路径
func (d *Disruptor[T]) reserveShared(count uint32) (int64, error) {
	if d.closed.Load() != stateRunning {
		return 0, ErrClosed
	}
	seq := d.shared.reserve(count)
	if seq >= 0 {
		if d.metrics != nil {
			d.metrics.ReserveCount(1)
		}
		return seq, nil
	}
	return 0, resolveError(seq)
}

// reserveSingleSlow 单写者慢路径：参数校验、关闭检查、消费者屏障刷新
func (d *Disruptor[T]) reserveSingleSlow(count uint32) (int64, error) {
	if count == 0 || count > d.capacity {
		return 0, ErrInvalidReservation
	}
	if d.closed.Load() != stateRunning {
		return 0, ErrClosed
	}

	slots := int64(count)
	d.reservedSequence += slots
	minimumSequence := d.reservedSequence - int64(d.capacity)

	// 自旋查询消费者屏障
	for spin := int64(0); ; spin++ {
		d.cachedConsumerSequence = d.singleConsumerBarrier.Load(0)
		if minimumSequence <= d.cachedConsumerSequence {
			break
		}
		d.singleWaiter.Reserve(spin)
	}

	// 刷新预计算剩余容量
	d.remainingCapacity = d.cachedConsumerSequence - d.reservedSequence + int64(d.capacity)
	if d.metrics != nil {
		d.metrics.ReserveCount(1)
		d.metrics.ReserveWaitCount(1)
	}
	return d.reservedSequence, nil
}

// TryReserve 非阻塞式尝试预留
func (d *Disruptor[T]) TryReserve(count uint32) (int64, error) {
	if d.shared != nil {
		if d.closed.Load() != stateRunning {
			return 0, ErrClosed
		}
		seq := d.shared.tryReserve(count)
		if seq >= 0 {
			return seq, nil
		}
		return 0, resolveError(seq)
	}
	// 单写者快路径
	slots := int64(count)
	if slots <= d.remainingCapacity {
		d.reservedSequence += slots
		d.remainingCapacity -= slots
		return d.reservedSequence, nil
	}
	return d.tryReserveSingleSlow(count)
}

// tryReserveSingleSlow 单写者 TryReserve 慢路径
func (d *Disruptor[T]) tryReserveSingleSlow(count uint32) (int64, error) {
	if count == 0 || count > d.capacity {
		return 0, ErrInvalidReservation
	}
	if d.closed.Load() != stateRunning {
		return 0, ErrClosed
	}

	slots := int64(count)
	d.reservedSequence += slots
	minimumSequence := d.reservedSequence - int64(d.capacity)

	// 单次刷新
	d.cachedConsumerSequence = d.singleConsumerBarrier.Load(0)
	if minimumSequence > d.cachedConsumerSequence {
		d.reservedSequence -= slots // 回滚
		return 0, ErrCapacityUnavailable
	}

	d.remainingCapacity = d.cachedConsumerSequence - d.reservedSequence + int64(d.capacity)
	return d.reservedSequence, nil
}

// Commit 使已预留的槽位对消费者可见。
// 单写者快路径：直接写入嵌入的 atomic.Int64，零指针追逐。
func (d *Disruptor[T]) Commit(lower, upper int64) {
	if d.shared != nil {
		d.shared.commit(lower, upper)
		if d.signal != nil {
			d.signal()
		}
		if d.metrics != nil {
			d.metrics.CommitCount(1)
		}
		return
	}
	// 单写者：通过预解引用指针直接 Store-Release
	d.committedStore.Store(upper)
	if d.signal != nil {
		d.signal()
	}
	if d.metrics != nil {
		d.metrics.CommitCount(1)
	}
}

// Listen 阻塞当前 goroutine，运行所有消费者 handler
func (d *Disruptor[T]) Listen() { d.listener.Listen() }

// Close 立即停止所有消费者，不等待排空
func (d *Disruptor[T]) Close() error {
	if !d.closed.CompareAndSwap(stateRunning, stateClosed) {
		return ErrClosed
	}
	// 毒化剩余容量：迫使下次 Reserve 进入慢路径（慢路径检查 closed 状态）
	d.remainingCapacity = -1
	return d.listener.Close()
}

// Drain 等待所有已提交事件被终端消费者处理完毕，然后停止
func (d *Disruptor[T]) Drain(ctx context.Context) error {
	if !d.closed.CompareAndSwap(stateRunning, stateClosed) {
		return ErrClosed
	}
	d.remainingCapacity = -1 // 毒化

	committed := d.committedSeq.Load()
	if committed == defaultSequenceValue {
		return d.listener.Close()
	}

	ticker := time.NewTicker(500 * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = d.listener.Close()
			return ctx.Err()
		case <-ticker.C:
			if d.terminalBarrier.Load(0) >= committed {
				return d.listener.Close()
			}
		}
	}
}

// resolveError 将 sentinel 负值转换为 Go error（冷路径）
func resolveError(sentinel int64) error {
	switch sentinel {
	case errReservationSize:
		return ErrInvalidReservation
	case errCapacityUnavailable:
		return ErrCapacityUnavailable
	default:
		return ErrClosed
	}
}
