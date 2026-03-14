// Metrics example: custom metrics collector for monitoring Disruptor internals
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gocronx/seqflow"
)

type Event struct{ Value int64 }

// simpleMetrics is an in-memory metrics collector.
// In production, replace with Prometheus, OpenTelemetry, etc.
type simpleMetrics struct {
	reserves  atomic.Int64
	commits   atomic.Int64
	waits     atomic.Int64
	handled   atomic.Int64
	events    atomic.Int64
	idleCount atomic.Int64
	gateCount atomic.Int64
}

func (m *simpleMetrics) ReserveCount(n int64)             { m.reserves.Add(n) }
func (m *simpleMetrics) CommitCount(n int64)              { m.commits.Add(n) }
func (m *simpleMetrics) ReserveWaitCount(n int64)         { m.waits.Add(n) }
func (m *simpleMetrics) HandleCount(_ string, n int64)    { m.handled.Add(n) }
func (m *simpleMetrics) HandleEvents(_ string, n int64)   { m.events.Add(n) }
func (m *simpleMetrics) IdleCount(_ string, n int64)      { m.idleCount.Add(n) }
func (m *simpleMetrics) GateCount(_ string, n int64)      { m.gateCount.Add(n) }
func (m *simpleMetrics) BufferUsage(used, capacity int64) {}

type nopHandler struct{}

func (nopHandler) Handle(int64, int64) {}

func main() {
	metrics := &simpleMetrics{}

	d, err := seqflow.New[Event](
		seqflow.WithCapacity(1024),
		seqflow.WithMetrics(metrics),
		seqflow.WithHandler("worker", nopHandler{}),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	// Send 10000 events
	rb := d.RingBuffer()
	total := 10000
	for i := 0; i < total; i++ {
		upper, _ := d.Reserve(1)
		rb.Set(upper, Event{Value: int64(i)})
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)

	fmt.Println("=== Disruptor Metrics ===")
	fmt.Printf("Reserve calls:      %d\n", metrics.reserves.Load())
	fmt.Printf("Commit calls:       %d\n", metrics.commits.Load())
	fmt.Printf("Slow-path waits:    %d\n", metrics.waits.Load())
	fmt.Printf("Handler batches:    %d\n", metrics.handled.Load())
	fmt.Printf("Total events:       %d\n", metrics.events.Load())
	fmt.Printf("Idle waits:         %d\n", metrics.idleCount.Load())
	fmt.Printf("Gate waits:         %d\n", metrics.gateCount.Load())
}
