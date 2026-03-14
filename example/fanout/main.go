// Fan-out example: one event dispatched to multiple independent consumers
//
//	Producer → [logger]
//	         → [monitor]
//	         → [archive]
//
// All three consumers process events independently in parallel.
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gocronx/seqflow"
)

type LogEvent struct {
	Level   string
	Message string
}

type namedCounter struct {
	name  string
	count atomic.Int64
}

func (c *namedCounter) Handle(lower, upper int64) {
	c.count.Add(upper - lower + 1)
}

func main() {
	logger := &namedCounter{name: "logger"}
	monitor := &namedCounter{name: "monitor"}
	archive := &namedCounter{name: "archive"}

	// Three handlers with no dependencies → same event processed by all three in parallel
	d, err := seqflow.New[LogEvent](
		seqflow.WithCapacity(1024),
		seqflow.WithHandler("logger", logger),
		seqflow.WithHandler("monitor", monitor),
		seqflow.WithHandler("archive", archive),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	rb := d.RingBuffer()
	messages := []LogEvent{
		{"INFO", "service started"},
		{"WARN", "connection timeout"},
		{"ERROR", "database error"},
		{"INFO", "reconnected"},
	}

	for _, msg := range messages {
		upper, _ := d.Reserve(1)
		rb.Set(upper, msg)
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)

	fmt.Printf("logger: %d, monitor: %d, archive: %d\n",
		logger.count.Load(), monitor.count.Load(), archive.count.Load())
}
