// Diamond DAG example:
//
//	Producer → [A decode] → [B risk]    → [D store]
//	                      → [C compute] ↗
//
// B and C run in parallel. D waits for both B and C to finish.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocronx/seqflow"
)

type TradeEvent struct {
	Symbol string
	Price  float64
	Qty    int64
}

type handler struct{ name string }

func (h *handler) Handle(lower, upper int64) {
	fmt.Printf("[%s] processed sequences %d..%d\n", h.name, lower, upper)
}

func main() {
	d, err := seqflow.New[TradeEvent](
		seqflow.WithCapacity(1024),
		seqflow.WithHandler("decode", &handler{"decode"}),
		seqflow.WithHandler("risk", &handler{"risk"}, seqflow.DependsOn("decode")),
		seqflow.WithHandler("compute", &handler{"compute"}, seqflow.DependsOn("decode")),
		seqflow.WithHandler("store", &handler{"store"}, seqflow.DependsOn("risk", "compute")),
	)
	if err != nil {
		panic(err)
	}

	go d.Listen()

	rb := d.RingBuffer()
	trades := []TradeEvent{
		{"AAPL", 178.5, 100},
		{"GOOG", 141.2, 50},
		{"TSLA", 245.8, 200},
	}

	for _, trade := range trades {
		upper, _ := d.Reserve(1)
		rb.Set(upper, trade)
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)

	fmt.Println("diamond pipeline done")
}
