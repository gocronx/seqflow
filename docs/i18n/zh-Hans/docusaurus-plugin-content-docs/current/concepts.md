---
sidebar_position: 3
---

# 核心概念

## RingBuffer[T]

泛型预分配环形缓冲区，容量必须是 2 的幂。

```go
rb, _ := seqflow.NewRingBuffer[MyEvent](1024)
rb.Set(seq, MyEvent{Value: 42})
event := rb.Get(seq) // 返回 *MyEvent（零拷贝指针）
```

`Get()` 返回指向缓冲区槽位的指针，该指针在生产者回绕覆写之前有效。

## Reserve / Commit

两阶段发布协议：

```go
upper, err := d.Reserve(1)     // 预留 1 个槽位
lower := upper                  // 单槽位时 lower == upper
d.RingBuffer().Set(lower, data) // 写入数据
d.Commit(lower, upper)          // 对消费者可见
```

批量发布：

```go
upper, _ := d.Reserve(16)       // 一次预留 16 个槽位
lower := upper - 15
for seq := lower; seq <= upper; seq++ {
    d.RingBuffer().Set(seq, data)
}
d.Commit(lower, upper)          // 16 个槽位一次性发布
```

批量预留摊销了原子操作开销。Reserve(16) 每条消息比 channel 快 ~160 倍。

## Handler

消费者回调接口：

```go
type Handler interface {
    Handle(lower, upper int64)
}
```

## 容量（Capacity）

`WithCapacity(n)` 设置环形缓冲区的槽位数量。

- 必须是 2 的幂（用位运算取模：`seq & (capacity-1)`）
- 太小：生产者频繁等待消费者
- 太大：浪费内存，影响缓存局部性
- 推荐：大多数场景 1024 ~ 65536

内存占用 = `n * sizeof(T)`。
