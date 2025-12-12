package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型（包含常用字段）
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UUIDModel 包含 UUID 的基础模型
type UUIDModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UUID      string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
