// Batch reserve example: claim multiple slots in one atomic operation
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocronx/seqflow"
)

type Event struct {
	Value int64
}

type batchPrinter struct{}

func (p *batchPrinter) Handle(lower, upper int64) {
	fmt.Printf("batch processed %d events [%d..%d]\n", upper-lower+1, lower, upper)
}

func main() {
	d, err := seqflow.New[Event](
		seqflow.WithCapacity(1024),
		seqflow.WithHandler("printer", &batchPrinter{}),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	rb := d.RingBuffer()

	// Reserve 16 slots at once
	batchSize := uint32(16)
	upper, _ := d.Reserve(batchSize)
	lower := upper - int64(batchSize) + 1

	for seq := lower; seq <= upper; seq++ {
		rb.Set(seq, Event{Value: seq * 10})
	}
	d.Commit(lower, upper)
	fmt.Printf("committed %d events\n", batchSize)

	// Another batch
	upper, _ = d.Reserve(8)
	lower = upper - 7
	for seq := lower; seq <= upper; seq++ {
		rb.Set(seq, Event{Value: seq * 10})
	}
	d.Commit(lower, upper)
	fmt.Println("committed 8 events")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)
}
