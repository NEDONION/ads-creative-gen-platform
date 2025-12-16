package service

import (
	"context"
	"fmt"
	"time"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/settings"
)

// Poller 控制轮询策略。
type Poller struct {
	Interval    time.Duration
	MaxAttempts int
	Sleep       func(time.Duration)
}

func (p *Poller) interval() time.Duration {
	if p.Interval <= 0 {
		return settings.PollInterval
	}
	return p.Interval
}

func (p *Poller) attempts() int {
	if p.MaxAttempts <= 0 {
		return settings.MaxPollAttempts
	}
	return p.MaxAttempts
}

func (p *Poller) sleep(d time.Duration) {
	if p.Sleep != nil {
		p.Sleep(d)
		return
	}
	time.Sleep(d)
}

// TaskProcessor 负责执行创意任务的完整工作流。
type TaskProcessor struct {
	llmClient     ports.TongyiClient
	storageClient ports.StorageUploader
	taskRepo      ports.TaskRepository
	assetRepo     ports.AssetRepository
	poller        Poller
}

// NewTaskProcessor 创建处理器，注入依赖与轮询策略。
func NewTaskProcessor(
	llmClient ports.TongyiClient,
	storageClient ports.StorageUploader,
	taskRepo ports.TaskRepository,
	assetRepo ports.AssetRepository,
	poller Poller,
) *TaskProcessor {
	return &TaskProcessor{
		llmClient:     llmClient,
		storageClient: storageClient,
		taskRepo:      taskRepo,
		assetRepo:     assetRepo,
		poller:        poller,
	}
}

// Process 执行任务，负责生成、轮询与落地。
func (p *TaskProcessor) Process(ctx context.Context, taskID uint) error {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now()
	if err := p.taskRepo.UpdateFields(ctx, taskID, map[string]interface{}{
		"status":     models.TaskProcessing,
		"started_at": now,
		"progress":   settings.ProgressStarted,
	}); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	task, err := p.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}

	plan := p.buildPlan(task)
	if len(plan) == 0 {
		return p.failTask(ctx, taskID, "无可执行的生成计划")
	}

	if !p.hasVariantPlan(task) {
		_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{"prompt_used": plan[0].Prompt})
	}

	_ = p.taskRepo.UpdateProgress(ctx, task.ID, settings.ProgressPrompted)

	var firstPublicURL string

	for idx, req := range plan {
		progressUpdater := p.makePollProgressUpdater(ctx, task.ID, idx, len(plan))
		result, err := p.runOne(ctx, task, req, progressUpdater)
		if err != nil {
			return p.failTask(ctx, task.ID, err.Error())
		}

		if idx == 0 && result.FirstPublicURL != "" {
			firstPublicURL = result.FirstPublicURL
			_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{"first_asset_url": firstPublicURL})
		}

		_ = p.taskRepo.UpdateProgress(ctx, task.ID, p.progressAfterRun(idx+1, len(plan)))
	}

	return p.completeTask(ctx, task.ID, now, firstPublicURL)
}

func (p *TaskProcessor) runOne(ctx context.Context, task *models.CreativeTask, req GenRequest, onPending func(int, int)) (GenResult, error) {
	queryResp, traceID, err := p.run(ctx, task, req, onPending)
	if err != nil {
		return GenResult{}, err
	}

	result, err := p.persistAssets(ctx, task, req, queryResp)
	if err != nil {
		p.finishTrace(traceID, "failed", "", err.Error())
		return GenResult{}, err
	}

	p.finishTrace(traceID, "success", result.FirstPublicURL, "")
	return result, nil
}

func (p *TaskProcessor) progressAfterRun(completedRuns int, totalRuns int) int {
	if totalRuns <= 0 {
		return settings.ProgressGenerated
	}
	return settings.ProgressGenerated + (completedRuns*(settings.ProgressCompleted-settings.ProgressGenerated))/totalRuns
}

func (p *TaskProcessor) makePollProgressUpdater(ctx context.Context, taskID uint, runIndex int, totalRuns int) func(int, int) {
	start := p.progressAfterRun(runIndex, totalRuns)
	end := p.progressAfterRun(runIndex+1, totalRuns)
	if end < start {
		end = start
	}

	return func(attempt int, attempts int) {
		if attempts <= 0 {
			return
		}

		progress := start
		if end > start {
			progress = start + (attempt*(end-start))/attempts
		}
		_ = p.taskRepo.UpdateProgress(ctx, taskID, progress)
	}
}

func (p *TaskProcessor) failTask(ctx context.Context, taskID uint, msg string) error {
	if msg == "" {
		msg = "任务失败，无具体错误信息"
	}
	_ = p.taskRepo.UpdateFields(ctx, taskID, map[string]interface{}{
		"status":        models.TaskFailed,
		"error_message": msg,
		"progress":      settings.ProgressCompleted,
	})
	return fmt.Errorf(msg)
}

func (p *TaskProcessor) completeTask(ctx context.Context, taskID uint, startedAt time.Time, firstURL string) error {
	completedAt := time.Now()
	duration := int(completedAt.Sub(startedAt).Seconds())

	update := map[string]interface{}{
		"status":              models.TaskCompleted,
		"progress":            settings.ProgressCompleted,
		"completed_at":        completedAt,
		"processing_duration": duration,
	}
	if firstURL != "" {
		update["first_asset_url"] = firstURL
	}

	return p.taskRepo.UpdateFields(ctx, taskID, update)
}
