// Basic example: single producer → single consumer
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocronx/seqflow"
)

type Order struct {
	ID    int64
	Price float64
}

// Order processor
type orderProcessor struct{}

func (p *orderProcessor) Handle(lower, upper int64) {
	fmt.Printf("processed order sequences %d..%d\n", lower, upper)
}

func main() {
	d, err := seqflow.New[Order](
		seqflow.WithCapacity(1024),
		seqflow.WithHandler("processor", &orderProcessor{}),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	// Publish 5 orders
	rb := d.RingBuffer()
	for i := int64(1); i <= 5; i++ {
		upper, _ := d.Reserve(1)
		rb.Set(upper, Order{ID: i, Price: float64(i) * 99.9})
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)

	fmt.Println("done")
}
