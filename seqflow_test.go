package seqflow

import (
	"context"
	"sync"
	"testing"
	"time"
)

type sumHandler struct {
	mu  sync.Mutex
	sum int64
}

func (h *sumHandler) Handle(lower, upper int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := lower; i <= upper; i++ {
		h.sum += i
	}
}

func TestNew_InvalidCapacity(t *testing.T) {
	h := &sumHandler{}
	_, err := New[int64](WithCapacity(0), WithHandler("a", h))
	if err == nil {
		t.Error("容量为 0 应返回错误")
	}
	_, err = New[int64](WithCapacity(3), WithHandler("a", h))
	if err == nil {
		t.Error("非 2 的幂应返回错误")
	}
}

func TestNew_NoHandlers(t *testing.T) {
	_, err := New[int64](WithCapacity(1024))
	if err == nil {
		t.Error("无 handler 应返回错误")
	}
}

func TestNew_DuplicateHandler(t *testing.T) {
	h := &sumHandler{}
	_, err := New[int64](WithCapacity(1024), WithHandler("a", h), WithHandler("a", h))
	if err == nil {
		t.Error("重复 handler 名应返回错误")
	}
}

func TestDisruptor_SingleProducerSingleConsumer(t *testing.T) {
	h := &sumHandler{}
	d, err := New[int64](
		WithCapacity(64),
		WithHandler("sum", h),
	)
	if err != nil {
		t.Fatalf("New 错误: %v", err)
	}

	go d.Listen()

	// 发布 10 个事件
	for i := int64(0); i < 10; i++ {
		upper, err := d.Reserve(1)
		if err != nil {
			t.Fatalf("Reserve 错误: %v", err)
		}
		d.RingBuffer().Set(upper, i*10)
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := d.Drain(ctx); err != nil {
		t.Fatalf("Drain 错误: %v", err)
	}

	// 序列 0..9 之和 = 45
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sum != 45 {
		t.Errorf("sum = %d, 期望 45（序列 0-9 之和）", h.sum)
	}
}

func TestDisruptor_MultiProducer(t *testing.T) {
	h := &sumHandler{}
	d, err := New[int64](
		WithCapacity(64),
		WithWriterCount(4),
		WithHandler("sum", h),
	)
	if err != nil {
		t.Fatalf("New 错误: %v", err)
	}

	go d.Listen()

	var wg sync.WaitGroup
	perWriter := 100
	wg.Add(4)
	for w := 0; w < 4; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				upper, err := d.Reserve(1)
				if err != nil {
					t.Errorf("Reserve 错误: %v", err)
					return
				}
				d.RingBuffer().Set(upper, 1)
				d.Commit(upper, upper)
			}
		}()
	}
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := d.Drain(ctx); err != nil {
		t.Fatalf("Drain 错误: %v", err)
	}
}

func TestDisruptor_DAG_Diamond(t *testing.T) {
	var aCount, bCount, cCount, dCount int64
	var mu sync.Mutex

	counter := func(ptr *int64) Handler {
		return handlerFunc(func(lower, upper int64) {
			mu.Lock()
			*ptr += upper - lower + 1
			mu.Unlock()
		})
	}

	d, err := New[int64](
		WithCapacity(64),
		WithHandler("A", counter(&aCount)),
		WithHandler("B", counter(&bCount), DependsOn("A")),
		WithHandler("C", counter(&cCount), DependsOn("A")),
		WithHandler("D", counter(&dCount), DependsOn("B", "C")),
	)
	if err != nil {
		t.Fatalf("New 错误: %v", err)
	}

	go d.Listen()

	for i := 0; i < 100; i++ {
		upper, _ := d.Reserve(1)
		d.RingBuffer().Set(upper, int64(i))
		d.Commit(upper, upper)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := d.Drain(ctx); err != nil {
		t.Fatalf("Drain 错误: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, tc := range []struct {
		name  string
		count int64
	}{
		{"A", aCount}, {"B", bCount}, {"C", cCount}, {"D", dCount},
	} {
		if tc.count != 100 {
			t.Errorf("handler %s 处理了 %d 个事件, 期望 100", tc.name, tc.count)
		}
	}
}

func TestDisruptor_Close(t *testing.T) {
	h := &sumHandler{}
	d, err := New[int64](WithCapacity(64), WithHandler("sum", h))
	if err != nil {
		t.Fatalf("New 错误: %v", err)
	}

	done := make(chan struct{})
	go func() {
		d.Listen()
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	d.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close 后 Listen 未返回")
	}
}

func TestDisruptor_ReserveAfterClose(t *testing.T) {
	h := &sumHandler{}
	d, _ := New[int64](WithCapacity(64), WithHandler("sum", h))
	go d.Listen()
	time.Sleep(10 * time.Millisecond)
	d.Close()

	_, err := d.Reserve(1)
	if err != ErrClosed {
		t.Errorf("Close 后 Reserve = %v, 期望 ErrClosed", err)
	}
	_, err = d.TryReserve(1)
	if err != ErrClosed {
		t.Errorf("Close 后 TryReserve = %v, 期望 ErrClosed", err)
	}
}

func TestDisruptor_DrainThenClose(t *testing.T) {
	h := &sumHandler{}
	d, _ := New[int64](WithCapacity(64), WithHandler("sum", h))
	go d.Listen()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = d.Drain(ctx)

	err := d.Close()
	if err != ErrClosed {
		t.Errorf("Drain 后 Close = %v, 期望 ErrClosed", err)
	}
}

func TestDisruptor_CloseThenDrain(t *testing.T) {
	h := &sumHandler{}
	d, _ := New[int64](WithCapacity(64), WithHandler("sum", h))
	go d.Listen()
	time.Sleep(10 * time.Millisecond)
	d.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := d.Drain(ctx)
	if err != ErrClosed {
		t.Errorf("Close 后 Drain = %v, 期望 ErrClosed", err)
	}
}

func TestDisruptor_WithWaitStrategy(t *testing.T) {
	h := &sumHandler{}
	d, err := New[int64](
		WithCapacity(64),
		WithWaitStrategy(NewYieldingStrategy()),
		WithHandler("sum", h),
	)
	if err != nil {
		t.Fatalf("New 错误: %v", err)
	}

	go d.Listen()
	upper, _ := d.Reserve(1)
	d.Commit(upper, upper)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d.Drain(ctx)
}
