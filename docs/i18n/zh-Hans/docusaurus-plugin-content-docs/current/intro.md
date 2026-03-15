---
sidebar_position: 1
---

# 简介

**seqflow** 是一个高性能、无锁的 Go Disruptor 库。

名称来自 **Seq**uence-driven **Flow**（序列驱动流）—— [LMAX Disruptor](https://github.com/LMAX-Exchange/disruptor) 模式的核心机制：所有协调通过序列号的原子推进完成，不用锁，不用 channel。

## 为什么用 seqflow

Go channel 在大多数场景下够用，但在高吞吐、低延迟场景下会成为瓶颈 —— 锁竞争、内存分配、GC 压力。

seqflow 用预分配的环形缓冲区 + 序列号协调替代 channel：

- 比 Go channel **快 ~10 倍**（单写者）
- 热路径**零内存分配**
- 支持 **DAG 消费者拓扑**

## 适用场景

- 金融交易系统（订单撮合、行情分发）
- 日志/事件管道（高吞吐采集 → 多级处理）
- 实时数据处理（IoT 传感器、监控聚合）
- 游戏服务器（帧同步、消息广播）

## 不适用场景

- 跨网络通信（用消息队列）
- 低频 CRUD（channel 够用）
- 需要持久化（用 Kafka）
