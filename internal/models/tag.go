package models

// Tag 标签
type Tag struct {
	BaseModel
	Name       string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Category   string `gorm:"type:varchar(64)" json:"category,omitempty"`
	Color      string `gorm:"type:varchar(7)" json:"color,omitempty"`
	UsageCount int    `gorm:"default:0" json:"usage_count"`

	// 关联
	Creatives []CreativeAsset `gorm:"many2many:creative_tags;" json:"creatives,omitempty"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}
