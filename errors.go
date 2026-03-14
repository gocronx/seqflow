package seqflow

import "errors"

var (
	ErrInvalidReservation  = errors.New("seqflow: invalid reservation size (zero or exceeds capacity)")
	ErrCapacityUnavailable = errors.New("seqflow: capacity unavailable")
	ErrClosed              = errors.New("seqflow: disruptor is closed")
	ErrInvalidCapacity     = errors.New("seqflow: capacity must be a positive power of 2")
	ErrNoHandlers          = errors.New("seqflow: at least one handler is required")
	ErrDuplicateHandler    = errors.New("seqflow: duplicate handler name")
	ErrUnknownDependency   = errors.New("seqflow: unknown dependency in DependsOn")
	ErrCyclicDependency    = errors.New("seqflow: cyclic dependency detected")
)

// Handler 是消费者回调接口，接收可用序列的批次范围
type Handler interface {
	Handle(lower, upper int64)
}
