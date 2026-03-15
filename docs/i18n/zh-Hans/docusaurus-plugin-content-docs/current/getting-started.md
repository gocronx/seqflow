---
sidebar_position: 2
---

# 快速开始

## 安装

```bash
go get github.com/gocronx/seqflow
```

需要 Go 1.22+。

## 基础示例

```go
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

type printer struct{}

func (p *printer) Handle(lower, upper int64) {
    fmt.Printf("处理序列 %d..%d\n", lower, upper)
}

func main() {
    d, err := seqflow.New[Event](
        seqflow.WithCapacity(1024),
        seqflow.WithHandler("printer", &printer{}),
    )
    if err != nil {
        panic(err)
    }

    go d.Listen()

    rb := d.RingBuffer()
    for i := int64(0); i < 10; i++ {
        upper, _ := d.Reserve(1)
        rb.Set(upper, Event{Value: i})
        d.Commit(upper, upper)
    }

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    d.Drain(ctx)
}
```

## 工作原理

1. **Reserve** — 生产者原子地预留环形缓冲区槽位
2. **Write** — 通过 `RingBuffer.Set()` 写入数据
3. **Commit** — 发布数据（原子 Store-Release）
4. **Handle** — 消费者收到序列范围 `(lower, upper)` 并处理

```mermaid
sequenceDiagram
    participant P as 生产者
    participant D as Disruptor
    participant RB as RingBuffer
    participant H as Handler

    P->>D: Reserve(n)
    D-->>P: 上界序列号
    P->>RB: rb.Set(seq, event)
    P->>D: Commit(lower, upper)
    D->>H: Handle(lower, upper)
    H->>RB: rb.Get(seq)
```
