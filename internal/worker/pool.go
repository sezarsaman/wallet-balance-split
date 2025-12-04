package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task تعریف یک کار که میتونه توسط worker pool انجام شود
type Task interface {
	Execute(ctx context.Context) error
}

// WorkerPool مدیریت کردن تعداد concurrent workers
type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorkerPool ایجاد یک worker pool جدید
func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, workers*2),
		ctx:       ctx,
		cancel:    cancel,
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	log.Printf("✅ Worker Pool initialized with %d workers", workers)
	return pool
}

// worker یک goroutine که tasks رو از queue پردازش میکند
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
		}
	}
}

// Submit یک task جدید به queue اضافه میکند
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

// SubmitWithTimeout تلاش میکند task رو با timeout submit کند
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

// Shutdown gracefully shutdown کردن worker pool
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

// GetQueueLength بازگرداندن طول queue (برای monitoring)
func (p *WorkerPool) GetQueueLength() int {
	return len(p.taskQueue)
}
