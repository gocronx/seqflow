---
sidebar_position: 7
---

# 性能

## 测试结果

Apple M4 / darwin arm64 / Go 1.22+ / 全部零分配。

### 单写者

| 场景 | seqflow | channel | 提升 |
|:-----|--------:|--------:|-----:|
| 每次 1 槽位 | **2.1 ns** | 21 ns | 10x |
| 每次 16 槽位 | **0.14 ns** | 22 ns/条 | 160x |

### 多写者（4 goroutine）

| 场景 | seqflow | channel | 提升 |
|:-----|--------:|--------:|-----:|
| 每次 1 槽位 | **39 ns** | 100 ns | 2.6x |
| 每次 16 槽位 | **2.3 ns** | 103 ns/条 | 45x |

## 运行 Benchmark

```bash
go test -bench=. -benchmem -count=3 -timeout 120s
```
