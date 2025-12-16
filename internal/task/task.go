package task

import "context"

// Task 抽象任务接口
type Task interface {
	Execute(ctx context.Context) error
	ID() uint
}
