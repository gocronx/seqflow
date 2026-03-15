---
sidebar_position: 1
---

# Introduction

**seqflow** is a high-performance, lock-free Disruptor library for Go.

The name comes from **Seq**uence-driven **Flow** — the core mechanism of the [LMAX Disruptor](https://github.com/LMAX-Exchange/disruptor) pattern: all coordination happens through atomic progression of sequence numbers, not locks or channels.

## Why seqflow

Go channels are great for most use cases. But in high-throughput, low-latency scenarios, they become the bottleneck — lock contention, memory allocation, GC pressure.

seqflow replaces channels with a pre-allocated ring buffer + sequence number coordination. The result:

- **~10x faster** than Go channel (single-writer)
- **Zero allocations** on the hot path
- **DAG consumer topology** for complex processing pipelines

## When to use

- Financial trading systems (order matching, market data)
- Log/event pipelines (high-throughput ingest → multi-stage processing)
- Real-time data processing (IoT sensors, monitoring aggregation)
- Game servers (frame sync, message broadcast)

## When NOT to use

- Cross-network communication (use a message queue)
- Low-frequency CRUD (channels are fine)
- Need persistence (use Kafka)
