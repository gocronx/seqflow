package seqflow

import "testing"

func TestNoopMetrics_DoesNotPanic(t *testing.T) {
	m := &NoopMetrics{}
	m.ReserveCount(1)
	m.CommitCount(1)
	m.ReserveWaitCount(1)
	m.HandleCount("test", 1)
	m.HandleEvents("test", 1)
	m.IdleCount("test", 1)
	m.GateCount("test", 1)
	m.BufferUsage(50, 100)
}

func TestNoopMetrics_ImplementsInterface(t *testing.T) {
	var _ Metrics = (*NoopMetrics)(nil)
}
