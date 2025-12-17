package service

import (
	"context"
	"fmt"
	"log"

	"ads-creative-gen-platform/internal/infra/llm"
	"ads-creative-gen-platform/internal/models"
)

func (p *TaskProcessor) run(ctx context.Context, task *models.CreativeTask, req GenRequest, onPending func(int, int)) (*llm.ImageGenResponse, string, error) {
	resp, traceID, err := p.submit(ctx, task, req)
	if err != nil {
		p.finishTrace(traceID, "failed", "", err.Error())
		return nil, traceID, err
	}

	queryResp, err := p.pollUntilDone(ctx, task.UUID, traceID, resp.Output.TaskID, onPending)
	if err != nil {
		p.finishTrace(traceID, "failed", "", err.Error())
		return nil, traceID, err
	}

	return queryResp, traceID, nil
}

func (p *TaskProcessor) submit(ctx context.Context, task *models.CreativeTask, req GenRequest) (*llm.ImageGenResponse, string, error) {
	numImages := req.NumImages
	if numImages <= 0 {
		numImages = 1
	}
	source := task.UUID
	if source == "" {
		source = task.ProductName
	}

	if task.ProductImageURL != "" {
		return p.llmClient.GenerateImageWithProduct(ctx, req.Prompt, task.ProductImageURL, req.Size, numImages, source, "", task.ProductName)
	}
	return p.llmClient.GenerateImage(ctx, req.Prompt, req.Size, numImages, source, "", task.ProductName)
}

func (p *TaskProcessor) pollUntilDone(
	ctx context.Context,
	requestID string,
	traceID string,
	tongyiTaskID string,
	onPending func(int, int),
) (*llm.ImageGenResponse, error) {
	attempts := p.poller.attempts()
	interval := p.poller.interval()

	for i := 0; i < attempts; i++ {
		p.poller.sleep(interval)

		queryResp, err := p.llmClient.QueryTask(ctx, traceID, tongyiTaskID, requestID)
		if err != nil {
			log.Printf("查询任务 %s 失败: %v", tongyiTaskID, err)
			continue
		}

		switch queryResp.Output.TaskStatus {
		case "SUCCEEDED":
			return queryResp, nil
		case "FAILED":
			msg := queryResp.Output.Message
			if msg == "" {
				msg = "任务失败，无具体错误信息"
			}
			return nil, fmt.Errorf(msg)
		default:
			if onPending != nil {
				onPending(i, attempts)
			}
		}
	}

	return nil, fmt.Errorf("任务在%d秒后超时", int(interval.Seconds())*attempts)
}

func (p *TaskProcessor) finishTrace(traceID, status, firstURL, errMsg string) {
	if traceID == "" {
		return
	}
	p.llmClient.FinishTrace(traceID, status, firstURL, errMsg)
}
