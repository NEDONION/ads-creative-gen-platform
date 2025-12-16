package settings

import "time"

// 任务处理配置
const (
	// DefaultImageSize 默认图片尺寸
	DefaultImageSize = "1024*1024"

	// DefaultImageWidth 默认图片宽度
	DefaultImageWidth = 1024

	// DefaultImageHeight 默认图片高度
	DefaultImageHeight = 1024

	// DefaultNumVariants 默认变体数量
	DefaultNumVariants = 2

	// DefaultFormat 默认格式
	DefaultFormat = "1:1"

	// ModelName 模型名称
	ModelName = "wanx-v1"
)

// 任务轮询配置
const (
	// MaxPollAttempts 最大轮询次数
	MaxPollAttempts = 60

	// PollInterval 轮询间隔
	PollInterval = 2 * time.Second

	// TaskTimeout 任务总超时时间
	TaskTimeout = 3 * time.Minute
)

// 进度常量
const (
	ProgressStart     = 0
	ProgressQueued    = 5
	ProgressStarted   = 10
	ProgressPrompted  = 30
	ProgressGenerated = 60
	ProgressCompleted = 100
)

// Worker Pool 配置
const (
	// DefaultWorkerCount 默认工作协程数
	DefaultWorkerCount = 10

	// DefaultQueueSize 默认队列大小
	DefaultQueueSize = 100
)
