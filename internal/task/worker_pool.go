package task

import (
	"ads-creative-gen-platform/internal/settings"
	"context"
	"fmt"
	"log"
	"sync"
)

// WorkerPool 工作池
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	running     bool
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = settings.DefaultWorkerCount
	}
	if queueSize <= 0 {
		queueSize = settings.DefaultQueueSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	if wp.running {
		wp.mu.Unlock()
		return
	}
	wp.running = true
	wp.mu.Unlock()

	log.Printf("Starting worker pool with %d workers", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	if !wp.running {
		wp.mu.Unlock()
		return
	}
	wp.running = false
	wp.mu.Unlock()

	log.Println("Stopping worker pool...")
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task Task) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.running {
		return fmt.Errorf("worker pool is not running")
	}

	select {
	case wp.taskQueue <- task:
		log.Printf("Task %d submitted to worker pool", task.ID())
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d recovered from panic: %v", id, r)
		}
	}()

	log.Printf("Worker %d started", id)

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue closed, exiting", id)
				return
			}

			log.Printf("Worker %d: processing task %d", id, task.ID())

			// 为每个任务创建独立的 context，设置超时
			taskCtx, cancel := context.WithTimeout(wp.ctx, settings.TaskTimeout)

			if err := task.Execute(taskCtx); err != nil {
				log.Printf("Worker %d: task %d failed: %v", id, task.ID(), err)
			} else {
				log.Printf("Worker %d: task %d completed successfully", id, task.ID())
			}

			cancel()

		case <-wp.ctx.Done():
			log.Printf("Worker %d: context cancelled, exiting", id)
			return
		}
	}
}

// IsRunning 检查工作池是否运行中
func (wp *WorkerPool) IsRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.running
}

// QueueLength 获取队列长度
func (wp *WorkerPool) QueueLength() int {
	return len(wp.taskQueue)
}
