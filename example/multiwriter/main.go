// Multi-writer example: 4 goroutines writing concurrently to one Disruptor
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocronx/seqflow"
)

type Event struct {
	ProducerID int
	Value      int64
}

// Counter handler that tracks total events received
type counter struct {
	total atomic.Int64
}

func (c *counter) Handle(lower, upper int64) {
	c.total.Add(upper - lower + 1)
}

func main() {
	c := &counter{}
	d, err := seqflow.New[Event](
		seqflow.WithCapacity(4096),
		seqflow.WithWriterCount(4), // enable multi-writer mode
		seqflow.WithHandler("counter", c),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	// 4 producers write concurrently
	var wg sync.WaitGroup
	perWriter := 1000
	wg.Add(4)

	for id := 0; id < 4; id++ {
		go func(pid int) {
			defer wg.Done()
			rb := d.RingBuffer()
			for i := 0; i < perWriter; i++ {
				upper, err := d.Reserve(1)
				if err != nil {
					return
				}
				rb.Set(upper, Event{ProducerID: pid, Value: int64(i)})
				d.Commit(upper, upper)
			}
		}(id)
	}

	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)

	fmt.Printf("4 producers sent %d events, consumer received %d\n", 4*perWriter, c.total.Load())
}
