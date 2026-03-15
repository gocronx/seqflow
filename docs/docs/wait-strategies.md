---
sidebar_position: 6
---

# Wait Strategies

Wait strategies control backpressure behavior for both producers and consumers.

```go
d, _ := seqflow.New[Event](
    seqflow.WithWaitStrategy(seqflow.NewYieldingStrategy()),
    // ...
)
```

## Available Strategies

| Strategy | Latency | CPU | Best For |
|----------|---------|-----|----------|
| `BusySpinStrategy` | Lowest | Highest | Dedicated CPU cores, ultra-low latency |
| `YieldingStrategy` | Low | Medium | Shared CPU, low latency requirements |
| `SleepingStrategy` | Medium | Low | **Default.** General purpose |
| `BlockingStrategy` | High | Lowest | Low-frequency, CPU-constrained environments |

## How They Work

Each strategy implements three callbacks:

- **Gate** — data is committed but upstream handler group hasn't finished yet (work is imminent)
- **Idle** — no data available anywhere
- **Reserve** — ring buffer is full, producer waiting for consumers

### BusySpinStrategy

Does nothing on all three callbacks. The goroutine spins continuously. Use only when you can dedicate a CPU core to the goroutine.

### YieldingStrategy

Calls `runtime.Gosched()` on all callbacks. Yields the processor to other goroutines but comes back quickly.

### SleepingStrategy (default)

- Gate: `runtime.Gosched()` (work is imminent, lightweight yield)
- Idle: `time.Sleep(500ns)`
- Reserve: `time.Sleep(1ns)`

Good balance between latency and CPU usage.

### BlockingStrategy

Uses `sync.Cond.Wait()` for all callbacks. Producer calls `Signal()` on Commit to wake consumers. Lowest CPU usage but highest latency.

## Choosing a Strategy

- Start with `SleepingStrategy` (default)
- If latency matters more than CPU: try `YieldingStrategy`
- If you have dedicated cores: try `BusySpinStrategy`
- If CPU matters more than latency: try `BlockingStrategy`
