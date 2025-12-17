package cache

import (
	"context"
	"time"
)

// Codec 定义序列化/反序列化行为，便于缓存任意结构
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// Cache 基础缓存接口（用于具体存储实现）
type Cache interface {
	Get(ctx context.Context, key string, dst any) (bool, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}

// DataCache 面向业务的高阶接口，封装多级缓存与单飞行
type DataCache interface {
	GetOrLoad(ctx context.Context, key string, ttl time.Duration, loader func(context.Context) (any, error), dst any) error
	Delete(ctx context.Context, key string)
	DeleteByPrefix(ctx context.Context, prefix string)
}
