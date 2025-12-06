package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

type Task interface {
	Execute(ctx context.Context) error
}

type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc

	done chan struct{}
}

func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, workers*2),
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}, workers*2),
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	log.Printf("✅ Worker Pool initialized with %d workers", workers)
	return pool
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("🛑 Worker %d shutting down", id)
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
			if err := task.Execute(ctx); err != nil {
				log.Printf("⚠️ Worker %d: Task execution failed: %v", id, err)
			}
			cancel()

			// NEW: signal task completed
			select {
			case p.done <- struct{}{}:
			default:
			}
		}
	}
}

func (p *WorkerPool) Wait(n int) {
	for i := 0; i < n; i++ {
		<-p.done
	}
}

func (p *WorkerPool) Submit(task Task) error {
	select {
	case <-p.ctx.Done():
		return ErrPoolClosed
	case p.taskQueue <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (p *WorkerPool) SubmitWithTimeout(ctx context.Context, task Task) error {
	select {
	case <-p.ctx.Done():
		return ErrPoolClosed
	case <-ctx.Done():
		return ctx.Err()
	case p.taskQueue <- task:
		return nil
	}
}

func (p *WorkerPool) Shutdown(timeout time.Duration) error {
	p.cancel()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Worker Pool shutdown successfully")
		return nil
	case <-time.After(timeout):
		log.Println("⚠️ Worker Pool shutdown timeout")
		return ErrShutdownTimeout
	}
}

func (p *WorkerPool) GetQueueLength() int {
	return len(p.taskQueue)
}
