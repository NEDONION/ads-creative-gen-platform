package task

import "fmt"

// Runner 统一封装 WorkerPool 的启动与任务提交
type Runner struct {
	wp *WorkerPool
}

// NewRunner 创建 runner
func NewRunner(workerCount, queueSize int) *Runner {
	return &Runner{wp: NewWorkerPool(workerCount, queueSize)}
}

// Start 启动 worker pool
func (r *Runner) Start() { r.wp.Start() }

// Stop 停止 worker pool
func (r *Runner) Stop() { r.wp.Stop() }

// Enqueue 提交任务
func (r *Runner) Enqueue(t Task) error {
	if r.wp == nil {
		return fmt.Errorf("worker pool not initialized")
	}
	return r.wp.Submit(t)
}

// QueueLength 返回当前队列长度
func (r *Runner) QueueLength() int {
	if r.wp == nil {
		return 0
	}
	return r.wp.QueueLength()
}
