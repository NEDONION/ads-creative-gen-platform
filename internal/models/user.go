package models

import (
	"time"
)

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

// UserStatus 用户状态
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned   UserStatus = "banned"
)

// User 用户模型
type User struct {
	UUIDModel
	Username     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string     `gorm:"type:varchar(20)" json:"phone,omitempty"`
	AvatarURL    string     `gorm:"type:varchar(512)" json:"avatar_url,omitempty"`
	Role         UserRole   `gorm:"type:enum('admin','user','viewer');default:'user'" json:"role"`
	Status       UserStatus `gorm:"type:enum('active','inactive','banned');default:'active'" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
