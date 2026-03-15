---
sidebar_position: 6
---

# 等待策略

| 策略 | 延迟 | CPU | 适用场景 |
|------|------|-----|---------|
| `BusySpinStrategy` | 极低 | 极高 | 独占 CPU 核心 |
| `YieldingStrategy` | 低 | 中 | 共享 CPU |
| `SleepingStrategy` | 中 | 低 | **默认**，通用场景 |
| `BlockingStrategy` | 高 | 极低 | 低频场景 |

```go
d, _ := seqflow.New[Event](
    seqflow.WithWaitStrategy(seqflow.NewYieldingStrategy()),
)
```

## 选择建议

- 默认用 `SleepingStrategy`
- 延迟优先：`YieldingStrategy`
- 独占核心：`BusySpinStrategy`
- CPU 优先：`BlockingStrategy`
