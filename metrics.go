package seqflow

// Metrics 收集可选的性能指标。当为 nil 时，热路径中通过 nil 检查跳过所有指标调用，零开销。
type Metrics interface {
	// ReserveCount 生产者 Reserve 调用次数
	ReserveCount(count int64)
	// CommitCount 生产者 Commit 调用次数
	CommitCount(count int64)
	// ReserveWaitCount Reserve 进入慢路径的次数
	ReserveWaitCount(count int64)
	// HandleCount 指定 handler 处理批次数
	HandleCount(name string, count int64)
	// HandleEvents 指定 handler 处理事件总数
	HandleEvents(name string, count int64)
	// IdleCount 指定 handler 空闲等待次数
	IdleCount(name string, count int64)
	// GateCount 指定 handler 被 gate 阻塞次数
	GateCount(name string, count int64)
	// BufferUsage 缓冲区使用率快照
	BufferUsage(used, capacity int64)
}

// NoopMetrics 是 Metrics 的空实现
type NoopMetrics struct{}

func (NoopMetrics) ReserveCount(int64)               {}
func (NoopMetrics) CommitCount(int64)                 {}
func (NoopMetrics) ReserveWaitCount(int64)            {}
func (NoopMetrics) HandleCount(string, int64)         {}
func (NoopMetrics) HandleEvents(string, int64)        {}
func (NoopMetrics) IdleCount(string, int64)           {}
func (NoopMetrics) GateCount(string, int64)           {}
func (NoopMetrics) BufferUsage(used, capacity int64)  {}
