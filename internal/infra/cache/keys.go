package cache

import (
	"fmt"

	"ads-creative-gen-platform/internal/shared"
)

// KeyBuilder 统一生成缓存 key，避免各业务重复拼接。
type KeyBuilder struct{}

func (KeyBuilder) TaskDetail(uuid string) string {
	return fmt.Sprintf("task:detail:%s", uuid)
}

func (KeyBuilder) TaskList(q shared.ListTasksQuery) string {
	return fmt.Sprintf("task:list:p=%d:ps=%d:status=%s:user=%d", q.Page, q.PageSize, q.Status, q.UserID)
}

func (KeyBuilder) AssetList(q shared.ListAssetsQuery) string {
	return fmt.Sprintf("asset:list:p=%d:ps=%d:fmt=%s:task=%s", q.Page, q.PageSize, q.Format, q.TaskID)
}

func (KeyBuilder) Experiment(uuid string) string {
	return fmt.Sprintf("exp:%s", uuid)
}

func (KeyBuilder) ExperimentList(status string, page, pageSize int) string {
	return fmt.Sprintf("explist:status=%s:p=%d:ps=%d", status, page, pageSize)
}

func (KeyBuilder) ExperimentMetrics(uuid string) string {
	return fmt.Sprintf("expmetrics:%s", uuid)
}

func (KeyBuilder) Asset(id uint) string {
	return fmt.Sprintf("asset:%d", id)
}

func (KeyBuilder) TraceList(status, modelName, traceID, productName string, page, pageSize int) string {
	return fmt.Sprintf("traces:list:s=%s:m=%s:id=%s:pname=%s:p=%d:ps=%d", status, modelName, traceID, productName, page, pageSize)
}

func (KeyBuilder) TraceDetail(traceID string) string {
	return fmt.Sprintf("traces:detail:%s", traceID)
}
