---
sidebar_position: 4
---

# API 参考

## 创建 Disruptor

```go
d, err := seqflow.New[T](opts ...Option) (*Disruptor[T], error)
```

### 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithCapacity(n)` | 1024 | 缓冲区大小（2 的幂） |
| `WithWriterCount(n)` | 1 | 并发写者数，>1 启用多写者模式 |
| `WithWaitStrategy(s)` | `SleepingStrategy` | 等待策略 |
| `WithMetrics(m)` | nil | 可选指标收集 |
| `WithHandler(name, h, opts...)` | 必填 | 注册命名 handler |

### 方法

| 方法 | 说明 |
|------|------|
| `Reserve(count) (int64, error)` | 预留槽位，返回上界序列号 |
| `TryReserve(count) (int64, error)` | 非阻塞版本 |
| `Commit(lower, upper)` | 发布数据 |
| `RingBuffer() *RingBuffer[T]` | 获取环形缓冲区 |
| `Listen()` | 阻塞运行消费者 |
| `Close() error` | 立即停止 |
| `Drain(ctx) error` | 排空后停止 |

### 错误

| 错误 | 说明 |
|------|------|
| `ErrInvalidReservation` | 槽位数为 0 或超过容量 |
| `ErrCapacityUnavailable` | TryReserve 时缓冲区已满 |
| `ErrClosed` | 已关闭 |
| `ErrNoHandlers` | 未注册 handler |
| `ErrDuplicateHandler` | handler 名称重复 |
| `ErrCyclicDependency` | 循环依赖 |
