package models

import "time"

// WarmupRecord 预热执行记录（持久化）
type WarmupRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StartedAt  time.Time `json:"started_at"`
	DurationMs int64     `json:"duration_ms"`
	Success    bool      `json:"success"`
	Actions    string    `json:"actions"` // JSON 数组字符串
	Errors     string    `json:"errors"`  // JSON 数组字符串
	CreatedAt  time.Time `json:"created_at"`
}

func (WarmupRecord) TableName() string {
	return "warmup_records"
}
