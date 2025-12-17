package cache

import (
	"time"

	"ads-creative-gen-platform/config"
)

// NewConfiguredCache 根据配置创建缓存（关闭时返回 Noop）
func NewConfiguredCache(cfg *config.Cache) DataCache {
	if cfg == nil || !cfg.Enabled {
		return NoopCache{}
	}
	local := NewLocalCache(cfg.MaxEntries, cfg.DefaultTTL, JSONCodec{})
	return NewLayeredCache(local, nil, JSONCodec{}, cfg.DefaultTTL)
}

// WithTTL helper 返回带覆写 TTL 的缓存引用（便于特定场景不同 TTL）
func WithTTL(base DataCache, ttl time.Duration) func() time.Duration {
	return func() time.Duration {
		return ttl
	}
}
