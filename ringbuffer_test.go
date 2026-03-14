package seqflow

import (
	"errors"
	"testing"
)

// TestNewRingBufferValidCapacity 验证有效容量能成功创建环形缓冲区
func TestNewRingBufferValidCapacity(t *testing.T) {
	for _, cap := range []uint32{1, 2, 4, 8, 16, 1024} {
		rb, err := NewRingBuffer[int64](cap)
		if err != nil {
			t.Errorf("容量 %d 应该合法，但得到错误: %v", cap, err)
		}
		if rb.Capacity() != cap {
			t.Errorf("期望容量 %d，实际得到 %d", cap, rb.Capacity())
		}
	}
}

// TestNewRingBufferInvalidCapacityZero 验证容量为 0 时返回错误
func TestNewRingBufferInvalidCapacityZero(t *testing.T) {
	_, err := NewRingBuffer[int64](0)
	if err == nil {
		t.Fatal("容量为 0 应该返回错误")
	}
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Errorf("期望 ErrInvalidCapacity，实际得到: %v", err)
	}
}

// TestNewRingBufferInvalidCapacityNonPowerOf2 验证非 2 的幂容量返回错误
func TestNewRingBufferInvalidCapacityNonPowerOf2(t *testing.T) {
	for _, cap := range []uint32{3, 5, 6, 7, 9, 10, 15, 100} {
		_, err := NewRingBuffer[int64](cap)
		if err == nil {
			t.Errorf("容量 %d 不是 2 的幂，应该返回错误", cap)
		}
		if !errors.Is(err, ErrInvalidCapacity) {
			t.Errorf("容量 %d: 期望 ErrInvalidCapacity，实际得到: %v", cap, err)
		}
	}
}

// TestRingBufferSetGet 验证基本的写入和读取操作
func TestRingBufferSetGet(t *testing.T) {
	rb, _ := NewRingBuffer[string](4)
	rb.Set(0, "hello")
	rb.Set(1, "world")

	if got := *rb.Get(0); got != "hello" {
		t.Errorf("期望 'hello'，实际得到 '%s'", got)
	}
	if got := *rb.Get(1); got != "world" {
		t.Errorf("期望 'world'，实际得到 '%s'", got)
	}
}

// TestRingBufferWrapAround 验证序列回绕后正确映射到槽位
func TestRingBufferWrapAround(t *testing.T) {
	rb, _ := NewRingBuffer[int](4) // 容量 4，掩码 3

	// 写入序列 0-3
	for i := int64(0); i < 4; i++ {
		rb.Set(i, int(i*10))
	}

	// 序列 4 应该回绕到槽位 0
	rb.Set(4, 40)
	if got := *rb.Get(4); got != 40 {
		t.Errorf("回绕后期望 40，实际得到 %d", got)
	}
	// 槽位 0 应该被覆写
	if got := *rb.Get(0); got != 40 {
		t.Errorf("槽位 0 应该被覆写为 40，实际得到 %d", got)
	}

	// 序列 7 回绕到槽位 3
	rb.Set(7, 70)
	if got := *rb.Get(7); got != 70 {
		t.Errorf("序列 7 回绕后期望 70，实际得到 %d", got)
	}
}

// TestRingBufferGetReturnsPointer 验证 Get 返回指针可直接修改槽位（零拷贝）
func TestRingBufferGetReturnsPointer(t *testing.T) {
	type Event struct {
		Value int
		Name  string
	}

	rb, _ := NewRingBuffer[Event](4)
	rb.Set(0, Event{Value: 1, Name: "original"})

	// 通过指针直接修改
	ptr := rb.Get(0)
	ptr.Value = 42
	ptr.Name = "modified"

	// 再次读取应该看到修改后的值
	if got := *rb.Get(0); got.Value != 42 || got.Name != "modified" {
		t.Errorf("通过指针修改后期望 {42, 'modified'}，实际得到 {%d, '%s'}", got.Value, got.Name)
	}
}
