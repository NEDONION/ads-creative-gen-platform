package models

// ProjectStatus 项目状态
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

// Project 项目/团队
type Project struct {
	UUIDModel
	Name        string        `gorm:"type:varchar(128);not null" json:"name"`
	Description string        `gorm:"type:text" json:"description,omitempty"`
	OwnerID     uint          `gorm:"not null;index" json:"owner_id"`
	Status      ProjectStatus `gorm:"type:enum('active','archived');default:'active';index" json:"status"`
	Settings    JSONMap       `gorm:"type:json" json:"settings,omitempty"`

	// 关联
	Owner   *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMemberRole 项目成员角色
type ProjectMemberRole string

const (
	ProjectRoleOwner  ProjectMemberRole = "owner"
	ProjectRoleAdmin  ProjectMemberRole = "admin"
	ProjectRoleMember ProjectMemberRole = "member"
	ProjectRoleViewer ProjectMemberRole = "viewer"
)

// ProjectMember 项目成员
type ProjectMember struct {
	BaseModel
	ProjectID uint              `gorm:"not null;index:idx_project_user" json:"project_id"`
	UserID    uint              `gorm:"not null;index:idx_project_user" json:"user_id"`
	Role      ProjectMemberRole `gorm:"type:enum('owner','admin','member','viewer');default:'member'" json:"role"`

	// 关联
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}
