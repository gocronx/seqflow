---
sidebar_position: 8
---

# 设计

## 架构

单包设计，无外部依赖。

## 核心优化

1. **预计算剩余容量** — 快路径 1 次比较
2. **零接口分发** — 单写者字段直接嵌入 Disruptor 结构体
3. **零原子读** — 关闭时毒化容量计数
4. **预解引用 Commit 指针** — 消除指针追逐
5. **缓存行对齐** — 防止 false sharing

## 单写者 vs 多写者

**SingleSequencer**（WriterCount=1）：快路径无原子操作。

**SharedSequencer**（WriterCount>1）：`atomic.Add` 抢占槽位 + 每槽位 round 追踪支持乱序提交。

## 优雅关闭

- `Close()` — 立即停止
- `Drain(ctx)` — 等待消费者追上已提交位置后关闭
