---
sidebar_position: 3
---

# Core Concepts

## RingBuffer[T]

Generic, pre-allocated ring buffer. Capacity must be a power of 2.

```go
rb, _ := seqflow.NewRingBuffer[MyEvent](1024)
rb.Set(seq, MyEvent{Value: 42})
event := rb.Get(seq) // returns *MyEvent (zero-copy pointer)
```

`Get()` returns a pointer directly into the buffer slot. The pointer is valid until the producer wraps around and overwrites that slot.

## Reserve / Commit

The two-phase publish protocol:

```go
upper, err := d.Reserve(1)     // claim 1 slot
lower := upper                  // for single slot, lower == upper
d.RingBuffer().Set(lower, data) // write data
d.Commit(lower, upper)          // make visible to consumers
```

For batch publishing:

```go
upper, _ := d.Reserve(16)       // claim 16 slots at once
lower := upper - 15
for seq := lower; seq <= upper; seq++ {
    d.RingBuffer().Set(seq, data)
}
d.Commit(lower, upper)          // one atomic publish for all 16
```

Batch reserve amortizes the atomic operation cost. Reserve(16) is ~160x faster per-message than channel.

## Handler

Consumer callback interface:

```go
type Handler interface {
    Handle(lower, upper int64)
}
```

The handler receives a range of sequences. Process them in order by reading from the ring buffer:

```go
func (h *MyHandler) Handle(lower, upper int64) {
    for seq := lower; seq <= upper; seq++ {
        event := ringBuffer.Get(seq)
        // process event
    }
}
```

## Capacity

`WithCapacity(n)` sets the number of slots in the ring buffer. Think of it as a circular conveyor belt with `n` slots.

- Must be a power of 2 (enables bitwise modulo: `seq & (capacity-1)`)
- Too small: producer blocks frequently waiting for consumers
- Too large: wastes memory, hurts cache locality
- Recommended: 1024 ~ 65536 for most use cases

Memory usage = `n * sizeof(T)`. For `int64` events with capacity 1024, that's 8 KB.
