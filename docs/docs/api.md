---
sidebar_position: 4
---

# API Reference

## Creating a Disruptor

```go
d, err := seqflow.New[T](opts ...Option) (*Disruptor[T], error)
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithCapacity(n)` | 1024 | Ring buffer size (power of 2) |
| `WithWriterCount(n)` | 1 | Concurrent writers. >1 enables shared sequencer |
| `WithWaitStrategy(s)` | `SleepingStrategy` | Backpressure strategy |
| `WithMetrics(m)` | nil | Optional metrics collector |
| `WithHandler(name, h, opts...)` | required | Register a named handler |

### Handler Options

| Option | Description |
|--------|-------------|
| `DependsOn(names...)` | Declare dependencies on other handlers |

## Disruptor Methods

### Reserve

```go
func (d *Disruptor[T]) Reserve(count uint32) (int64, error)
```

Claims `count` slots in the ring buffer. Returns the upper sequence number. Blocks if buffer is full.

### TryReserve

```go
func (d *Disruptor[T]) TryReserve(count uint32) (int64, error)
```

Non-blocking version. Returns `ErrCapacityUnavailable` if buffer is full.

### Commit

```go
func (d *Disruptor[T]) Commit(lower, upper int64)
```

Makes reserved slots visible to consumers. Must be called exactly once after Reserve.

### RingBuffer

```go
func (d *Disruptor[T]) RingBuffer() *RingBuffer[T]
```

Returns the ring buffer for reading/writing events.

### Listen

```go
func (d *Disruptor[T]) Listen()
```

Blocks the calling goroutine, running all consumer handlers. Call with `go d.Listen()`.

### Close

```go
func (d *Disruptor[T]) Close() error
```

Immediately stops all consumers without draining.

### Drain

```go
func (d *Disruptor[T]) Drain(ctx context.Context) error
```

Waits for all committed events to be processed, then stops. Respects context cancellation.

`Close` and `Drain` are mutually exclusive. Second call returns `ErrClosed`.

## Errors

| Error | Description |
|-------|-------------|
| `ErrInvalidReservation` | Zero slots or exceeds capacity |
| `ErrCapacityUnavailable` | TryReserve: buffer full |
| `ErrClosed` | Disruptor has been shut down |
| `ErrInvalidCapacity` | Capacity not a positive power of 2 |
| `ErrNoHandlers` | No handlers registered |
| `ErrDuplicateHandler` | Duplicate handler name |
| `ErrUnknownDependency` | DependsOn references unknown handler |
| `ErrCyclicDependency` | Circular dependency detected |
