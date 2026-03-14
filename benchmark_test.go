package seqflow

import (
	"sync"
	"testing"
	"time"
)

const (
	benchBufferSize = 1 << 16 // 64K
	benchReserve1   = 1
	benchReserve16  = 16
)

// === seqflow 基准测试 ===

func BenchmarkSeqflow_SingleWriter_Reserve1(b *testing.B) {
	benchmarkSeqflow(b, benchReserve1, 1, 1)
}

func BenchmarkSeqflow_SingleWriter_Reserve16(b *testing.B) {
	benchmarkSeqflow(b, benchReserve16, 1, 1)
}

func BenchmarkSeqflow_MultiWriter4_Reserve1(b *testing.B) {
	benchmarkSeqflow(b, benchReserve1, 4, 1)
}

func BenchmarkSeqflow_MultiWriter4_Reserve16(b *testing.B) {
	benchmarkSeqflow(b, benchReserve16, 4, 1)
}

func benchmarkSeqflow(b *testing.B, count uint32, writerCount int, consumerCount int) {
	handlers := make([]Option, consumerCount)
	for i := range handlers {
		name := string(rune('A' + i))
		handlers[i] = WithHandler(name, nopHandler{})
	}

	opts := append([]Option{
		WithCapacity(benchBufferSize),
		WithWriterCount(uint8(writerCount)),
	}, handlers...)

	d, err := New[int64](opts...)
	if err != nil {
		b.Fatalf("New 错误: %v", err)
	}
	defer d.Listen()

	go func() {
		var wg sync.WaitGroup
		wg.Add(writerCount)
		defer func() { wg.Wait(); _ = d.Close() }()
		time.Sleep(100 * time.Millisecond)
		b.ReportAllocs()
		b.ResetTimer()

		iterations := int64(b.N)
		slots := int64(count)
		offset := slots - 1

		for i := 0; i < writerCount; i++ {
			go func() {
				defer wg.Done()
				for seq := int64(defaultSequenceValue); seq < iterations; {
					upper, err := d.Reserve(count)
					if err != nil {
						return
					}
					seq = upper
					d.Commit(upper-offset, upper)
				}
			}()
		}
	}()
}

// === Channel 基准测试 ===

func BenchmarkChannel_SingleWriter(b *testing.B) {
	benchmarkChannel(b, 1)
}

func BenchmarkChannel_SingleWriter_Batch16(b *testing.B) {
	benchmarkChannelBatch(b, 1, 16)
}

func BenchmarkChannel_MultiWriter4(b *testing.B) {
	benchmarkChannel(b, 4)
}

func BenchmarkChannel_MultiWriter4_Batch16(b *testing.B) {
	benchmarkChannelBatch(b, 4, 16)
}

// benchmarkChannelBatch 每次迭代发送 batch 条消息，报告每条消息的摊销耗时
func benchmarkChannelBatch(b *testing.B, writers int, batch int) {
	ch := make(chan int64, benchBufferSize)
	iterations := int64(b.N)
	batchSize := int64(batch)
	totalPerWriter := iterations * batchSize

	b.ReportAllocs()
	b.ResetTimer()

	for w := 0; w < writers; w++ {
		go func() {
			for i := int64(0); i < totalPerWriter; i++ {
				ch <- i
			}
		}()
	}

	for i := int64(0); i < totalPerWriter*int64(writers); i++ {
		<-ch
	}
}

func benchmarkChannel(b *testing.B, writers int) {
	ch := make(chan int64, benchBufferSize)
	iterations := int64(b.N)

	b.ReportAllocs()
	b.ResetTimer()

	for w := 0; w < writers; w++ {
		go func() {
			for i := int64(0); i < iterations; i++ {
				ch <- i
			}
		}()
	}

	for i := int64(0); i < iterations*int64(writers); i++ {
		<-ch
	}
}

// === WaitStrategy 基准测试 ===

func BenchmarkWaitStrategy_BusySpin(b *testing.B) {
	benchmarkWithStrategy(b, NewBusySpinStrategy())
}

func BenchmarkWaitStrategy_Yielding(b *testing.B) {
	benchmarkWithStrategy(b, NewYieldingStrategy())
}

func BenchmarkWaitStrategy_Sleeping(b *testing.B) {
	benchmarkWithStrategy(b, NewSleepingStrategy())
}

// 注意：BlockingStrategy 使用 sync.Cond，与 defer Listen 的 benchmark 模式不兼容，
// 因为 producer 会在 consumer 启动前阻塞在 Cond.Wait 上导致死锁。
// BlockingStrategy 适用于实际应用中 Listen 先启动的场景。

func benchmarkWithStrategy(b *testing.B, ws WaitStrategy) {
	d, err := New[int64](
		WithCapacity(benchBufferSize),
		WithWaitStrategy(ws),
		WithHandler("A", nopHandler{}),
	)
	if err != nil {
		b.Fatalf("New 错误: %v", err)
	}
	defer d.Listen()

	go func() {
		time.Sleep(100 * time.Millisecond)
		b.ReportAllocs()
		b.ResetTimer()

		for i := int64(0); i < int64(b.N); i++ {
			upper, err := d.Reserve(1)
			if err != nil {
				return
			}
			d.Commit(upper, upper)
		}
		_ = d.Close()
	}()
}

// === DAG 拓扑基准测试 ===

func BenchmarkDAG_Linear_3Stage(b *testing.B) {
	d, _ := New[int64](
		WithCapacity(benchBufferSize),
		WithHandler("A", nopHandler{}),
		WithHandler("B", nopHandler{}, DependsOn("A")),
		WithHandler("C", nopHandler{}, DependsOn("B")),
	)
	benchmarkDisruptorRun(b, d, 1, 1)
}

func BenchmarkDAG_Diamond_4Handler(b *testing.B) {
	d, _ := New[int64](
		WithCapacity(benchBufferSize),
		WithHandler("A", nopHandler{}),
		WithHandler("B", nopHandler{}, DependsOn("A")),
		WithHandler("C", nopHandler{}, DependsOn("A")),
		WithHandler("D", nopHandler{}, DependsOn("B", "C")),
	)
	benchmarkDisruptorRun(b, d, 1, 1)
}

func BenchmarkDAG_FanOut_4Handler(b *testing.B) {
	d, _ := New[int64](
		WithCapacity(benchBufferSize),
		WithHandler("A", nopHandler{}),
		WithHandler("B", nopHandler{}, DependsOn("A")),
		WithHandler("C", nopHandler{}, DependsOn("A")),
		WithHandler("D", nopHandler{}, DependsOn("A")),
	)
	benchmarkDisruptorRun(b, d, 1, 1)
}

func benchmarkDisruptorRun(b *testing.B, d *Disruptor[int64], count uint32, writerCount int) {
	defer d.Listen()

	go func() {
		var wg sync.WaitGroup
		wg.Add(writerCount)
		defer func() { wg.Wait(); _ = d.Close() }()
		time.Sleep(100 * time.Millisecond)
		b.ReportAllocs()
		b.ResetTimer()

		iterations := int64(b.N)
		slots := int64(count)
		offset := slots - 1

		for i := 0; i < writerCount; i++ {
			go func() {
				defer wg.Done()
				for seq := int64(defaultSequenceValue); seq < iterations; {
					upper, err := d.Reserve(count)
					if err != nil {
						return
					}
					seq = upper
					d.Commit(upper-offset, upper)
				}
			}()
		}
	}()
}

// nopHandler 空操作 handler，用于基准测试
type nopHandler struct{}

func (nopHandler) Handle(int64, int64) {}
