---
sidebar_position: 7
---

# Performance

## Benchmark Results

Apple M4 / darwin arm64 / Go 1.22+ / 0 allocs across all tests.

### Single Writer

| Scenario | seqflow | channel | Speedup |
|:---------|--------:|--------:|--------:|
| 1 slot per op | **2.1 ns** | 21 ns | 10x |
| 16 slots per op | **0.14 ns** | 22 ns/msg | 160x |

### Multi Writer (4 goroutines)

| Scenario | seqflow | channel | Speedup |
|:---------|--------:|--------:|--------:|
| 1 slot per op | **39 ns** | 100 ns | 2.6x |
| 16 slots per op | **2.3 ns** | 103 ns/msg | 45x |

### Why is batch so fast?

`Reserve(16)` claims 16 slots with a single atomic operation. Channel must send 16 times, each with a lock acquisition.

## Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem -count=3 -timeout 120s

# Single writer only
go test -bench=BenchmarkSeqflow_SingleWriter -benchmem

# WaitStrategy comparison
go test -bench=BenchmarkWaitStrategy -benchmem

# DAG topology
go test -bench=BenchmarkDAG -benchmem
```

## Design Optimizations

1. **Pre-computed remaining capacity** — fast path: 1 comparison instead of subtraction + 2 comparisons
2. **Zero interface dispatch** — single-writer fields embedded directly in Disruptor struct
3. **Zero atomic loads on fast path** — Close poisons capacity counter
4. **Pre-dereferenced commit pointer** — eliminates pointer chasing in Commit
5. **Cache-line padding** — atomic sequences aligned to CPU cache lines, prevents false sharing
