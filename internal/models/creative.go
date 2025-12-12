package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskQueued     TaskStatus = "queued"
	TaskProcessing TaskStatus = "processing"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
	TaskCancelled  TaskStatus = "cancelled"
)

// StringArray 字符串数组类型（用于 JSON 存储）
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// CreativeTask 创意生成任务
type CreativeTask struct {
	UUIDModel
	UserID    uint  `gorm:"not null;index" json:"user_id"`
	ProjectID *uint `gorm:"index" json:"project_id,omitempty"`

	// 输入信息
	Title           string      `gorm:"type:varchar(255);not null" json:"title"`
	SellingPoints   StringArray `gorm:"type:json" json:"selling_points"`
	ProductImageURL string      `gorm:"type:varchar(512)" json:"product_image_url,omitempty"`
	BrandLogoURL    string      `gorm:"type:varchar(512)" json:"brand_logo_url,omitempty"`

	// 生成配置
	RequestedFormats StringArray `gorm:"type:json" json:"requested_formats"`
	RequestedStyles  StringArray `gorm:"type:json" json:"requested_styles"`
	NumVariants      int         `gorm:"default:3" json:"num_variants"`
	CTAText          string      `gorm:"type:varchar(64)" json:"cta_text,omitempty"`

	// 任务状态
	Status       TaskStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	Progress     int        `gorm:"default:0" json:"progress"` // 0-100
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`

	// 时间统计
	QueuedAt           *time.Time `json:"queued_at,omitempty"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	ProcessingDuration *int       `json:"processing_duration,omitempty"` // 秒

	// 关联
	User    *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Assets  []CreativeAsset `gorm:"foreignKey:TaskID" json:"assets,omitempty"`
}

// TableName 指定表名
func (CreativeTask) TableName() string {
	return "creative_tasks"
}

// StorageType 存储类型
type StorageType string

const (
	StorageLocal StorageType = "local"
	StorageOSS   StorageType = "oss"
	StorageS3    StorageType = "s3"
	StorageMinio StorageType = "minio"
	StorageQiniu StorageType = "qiniu"
)

// JSONMap 用于存储 JSON 对象
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// CreativeAsset 创意素材
type CreativeAsset struct {
	UUIDModel
	TaskID uint `gorm:"not null;index" json:"task_id"`

	// 素材信息
	Format   string `gorm:"type:varchar(20);not null;index" json:"format"`
	Width    int    `gorm:"not null" json:"width"`
	Height   int    `gorm:"not null" json:"height"`
	FileSize *int   `json:"file_size,omitempty"`

	// 存储信息
	StorageType  StorageType `gorm:"type:varchar(20);default:'local';not null" json:"storage_type"`
	PublicURL    string      `gorm:"type:varchar(1024);not null" json:"public_url"`    // 公共访问URL
	OriginalPath string      `gorm:"type:varchar(512)" json:"original_path,omitempty"` // 原始内部路径

	// 生成元数据
	Style            string  `gorm:"type:varchar(64);index" json:"style,omitempty"`
	VariantIndex     *int    `json:"variant_index,omitempty"`
	GenerationPrompt string  `gorm:"type:text" json:"generation_prompt,omitempty"`
	ModelName        string  `gorm:"type:varchar(128)" json:"model_name,omitempty"`
	ModelVersion     string  `gorm:"type:varchar(64)" json:"model_version,omitempty"`
	GenerationParams JSONMap `gorm:"type:json" json:"generation_params,omitempty"`

	// 内容信息
	TextContent JSONMap `gorm:"type:json" json:"text_content,omitempty"`
	HasLogo     bool    `gorm:"default:false" json:"has_logo"`
	HasCTA      bool    `gorm:"default:false" json:"has_cta"`

	// 排序
	Rank *int `gorm:"index" json:"rank,omitempty"`

	// 关联
	Task  *CreativeTask  `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Score *CreativeScore `gorm:"foreignKey:CreativeID" json:"score,omitempty"`
	Tags  []Tag          `gorm:"many2many:creative_tags;" json:"tags,omitempty"`
}

// TableName 指定表名
func (CreativeAsset) TableName() string {
	return "creative_assets"
}

// CreativeScore 创意评分
type CreativeScore struct {
	ID         uint `gorm:"primarykey" json:"id"`
	CreativeID uint `gorm:"uniqueIndex;not null" json:"creative_id"`

	// 质量评分
	QualityOverall    *float64 `gorm:"type:decimal(4,3)" json:"quality_overall,omitempty"`
	BrightnessScore   *float64 `gorm:"type:decimal(4,3)" json:"brightness_score,omitempty"`
	ContrastScore     *float64 `gorm:"type:decimal(4,3)" json:"contrast_score,omitempty"`
	SharpnessScore    *float64 `gorm:"type:decimal(4,3)" json:"sharpness_score,omitempty"`
	CompositionScore  *float64 `gorm:"type:decimal(4,3)" json:"composition_score,omitempty"`
	ColorHarmonyScore *float64 `gorm:"type:decimal(4,3)" json:"color_harmony_score,omitempty"`

	// CTR 预测
	CTRPrediction *float64 `gorm:"type:decimal(5,4);index" json:"ctr_prediction,omitempty"`
	CTRConfidence *float64 `gorm:"type:decimal(4,3)" json:"ctr_confidence,omitempty"`
	ModelVersion  string   `gorm:"type:varchar(64)" json:"model_version,omitempty"`

	// CLIP 评分
	CLIPScore      *float64 `gorm:"type:decimal(4,3)" json:"clip_score,omitempty"`
	AestheticScore *float64 `gorm:"type:decimal(4,3)" json:"aesthetic_score,omitempty"`

	// 安全检测
	NSFWScore *float64 `gorm:"type:decimal(4,3)" json:"nsfw_score,omitempty"`
	IsSafe    bool     `gorm:"default:true" json:"is_safe"`

	ScoredAt  time.Time `json:"scored_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Creative *CreativeAsset `gorm:"foreignKey:CreativeID" json:"creative,omitempty"`
}

// TableName 指定表名
func (CreativeScore) TableName() string {
	return "creative_scores"
}
