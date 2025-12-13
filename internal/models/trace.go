package models

import "time"

// ModelTrace 记录模型调用链路
type ModelTrace struct {
	ID            uint             `gorm:"primarykey" json:"id"`
	TraceID       string           `gorm:"type:varchar(64);uniqueIndex;not null" json:"trace_id"`
	ModelName     string           `gorm:"type:varchar(128);index" json:"model_name"`
	ModelVersion  string           `gorm:"type:varchar(64)" json:"model_version"`
	Status        string           `gorm:"type:varchar(16);index" json:"status"` // success/failed/running
	DurationMs    int              `json:"duration_ms"`
	StartAt       time.Time        `json:"start_at"`
	EndAt         time.Time        `json:"end_at"`
	Source        string           `gorm:"type:varchar(255)" json:"source,omitempty"` // 可放 experiment/task/user
	InputPreview  string           `gorm:"type:text" json:"input_preview,omitempty"`
	OutputPreview string           `gorm:"type:text" json:"output_preview,omitempty"`
	ErrorMessage  string           `gorm:"type:text" json:"error_message,omitempty"`
	Steps         []ModelTraceStep `gorm:"foreignKey:TraceID;references:TraceID" json:"steps,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// ModelTraceStep 记录链路步骤
type ModelTraceStep struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	TraceID       string    `gorm:"type:varchar(64);index" json:"trace_id"`
	StepName      string    `gorm:"type:varchar(128)" json:"step_name"`
	Component     string    `gorm:"type:varchar(128)" json:"component"`
	Status        string    `gorm:"type:varchar(16);index" json:"status"`
	DurationMs    int       `json:"duration_ms"`
	StartAt       time.Time `json:"start_at"`
	EndAt         time.Time `json:"end_at"`
	InputPreview  string    `gorm:"type:text" json:"input_preview,omitempty"`
	OutputPreview string    `gorm:"type:text" json:"output_preview,omitempty"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	Extra         JSONMap   `gorm:"type:json" json:"extra,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ModelTrace) TableName() string {
	return "model_traces"
}

func (ModelTraceStep) TableName() string {
	return "model_trace_steps"
}
