package tracing

import (
	"context"
	"time"
)

type traceKeyType string

const traceKey traceKeyType = "model_trace_id"

// 导出 context key 供复用
var CtxKeyTraceID traceKeyType = traceKey

// Tracer 提供无侵入的链路记录封装，基于 context 传递 trace_id。
type Tracer struct {
	svc *TraceService
}

func NewTracer() *Tracer {
	return &Tracer{svc: NewTraceService()}
}

// Start 开始一条 trace，返回携带 trace_id 的 context。
func (t *Tracer) Start(ctx context.Context, modelName, modelVersion, source, inputPreview, productName string) (context.Context, string) {
	traceID, err := t.svc.StartTrace(modelName, modelVersion, source, inputPreview, productName)
	if err != nil {
		return ctx, ""
	}
	ctx = context.WithValue(ctx, traceKey, traceID)
	return ctx, traceID
}

// Finish 结束 trace
func (t *Tracer) Finish(ctx context.Context, status, outputPreview, errorMessage string) {
	traceID := FromContext(ctx)
	if traceID == "" {
		return
	}
	_ = t.svc.FinishTrace(traceID, status, outputPreview, errorMessage)
}

// Step 记录步骤
func (t *Tracer) Step(ctx context.Context, stepName, component, status, inputPreview, outputPreview, errorMessage string, startAt, endAt time.Time) {
	traceID := FromContext(ctx)
	if traceID == "" {
		return
	}
	_ = t.svc.AddStep(traceID, stepName, component, status, inputPreview, outputPreview, errorMessage, startAt, endAt)
}

// FromContext 读取 trace_id
func FromContext(ctx context.Context) string {
	val := ctx.Value(traceKey)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// FinishTrace 便捷方法，供外部调用
func (t *Tracer) FinishTrace(traceID, status, outputPreview, errorMessage string) {
	if traceID == "" {
		return
	}
	_ = t.svc.FinishTrace(traceID, status, outputPreview, errorMessage)
}
