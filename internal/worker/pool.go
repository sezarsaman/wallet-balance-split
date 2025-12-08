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
	workers     int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	taskTimeout time.Duration
}

func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	p := &WorkerPool{
		workers:     workers,
		taskQueue:   make(chan Task, workers*4),
		ctx:         ctx,
		cancel:      cancel,
		taskTimeout: 30 * time.Second,
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.runWorker(i)
	}

	log.Printf("Worker Pool started with %d workers", workers)
	return p
}

func (p *WorkerPool) runWorker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return

		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			ctx, cancel := context.WithTimeout(p.ctx, p.taskTimeout)
			err := task.Execute(ctx)
			cancel()

			if err != nil {
				log.Printf("Worker %d error: %v", id, err)
			}
		}
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

func (p *WorkerPool) Shutdown(timeout time.Duration) error {
	log.Println("Worker pool shutting down...")
	p.cancel()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stopped")
		return nil
	case <-time.After(timeout):
		return ErrShutdownTimeout
	}
}

func (p *WorkerPool) GetQueueLength() int {
	return len(p.taskQueue)
}
