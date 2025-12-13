package models

import "time"

// ExperimentStatus 实验状态
type ExperimentStatus string

const (
	ExpDraft    ExperimentStatus = "draft"
	ExpActive   ExperimentStatus = "active"
	ExpPaused   ExperimentStatus = "paused"
	ExpArchived ExperimentStatus = "archived"
)

// Experiment 实验
type Experiment struct {
	UUIDModel
	Name        string              `gorm:"type:varchar(128);not null" json:"name"`
	ProductName string              `gorm:"type:varchar(255)" json:"product_name,omitempty"`
	Status      ExperimentStatus    `gorm:"type:varchar(16);default:'draft';index" json:"status"`
	StartAt     *time.Time          `json:"start_at,omitempty"`
	EndAt       *time.Time          `json:"end_at,omitempty"`
	Variants    []ExperimentVariant `gorm:"foreignKey:ExperimentID" json:"variants,omitempty"`
}

// ExperimentVariant 实验变体
type ExperimentVariant struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	ExperimentID uint      `gorm:"not null;index:idx_exp_creative" json:"experiment_id"`
	CreativeID   uint      `gorm:"not null;index:idx_exp_creative" json:"creative_id"`
	Weight       float64   `gorm:"type:decimal(6,3);default:0" json:"weight"` // 0-1
	BucketStart  int       `gorm:"not null" json:"bucket_start"`              // 0-10000
	BucketEnd    int       `gorm:"not null" json:"bucket_end"`                // 0-10000
	Title        string    `gorm:"type:varchar(255)" json:"title,omitempty"`
	ProductName  string    `gorm:"type:varchar(255)" json:"product_name,omitempty"`
	ImageURL     string    `gorm:"type:varchar(1024)" json:"image_url,omitempty"`
	CTAText      string    `gorm:"type:varchar(128)" json:"cta_text,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ExperimentMetric 实验指标
type ExperimentMetric struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	ExperimentID uint      `gorm:"not null;index:idx_exp_metric" json:"experiment_id"`
	CreativeID   uint      `gorm:"not null;index:idx_exp_metric" json:"creative_id"`
	Impressions  int64     `gorm:"default:0" json:"impressions"`
	Clicks       int64     `gorm:"default:0" json:"clicks"`
	CTR          *float64  `gorm:"type:decimal(8,6)" json:"ctr,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Experiment) TableName() string {
	return "experiments"
}

func (ExperimentVariant) TableName() string {
	return "experiment_variants"
}

func (ExperimentMetric) TableName() string {
	return "experiment_metrics"
}
