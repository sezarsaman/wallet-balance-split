package worker

import "errors"

var (
	ErrPoolClosed      = errors.New("worker pool is closed")
	ErrQueueFull       = errors.New("task queue is full")
	ErrShutdownTimeout = errors.New("shutdown timeout exceeded")
)
