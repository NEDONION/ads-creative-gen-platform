package cache

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// NoopCache 在关闭缓存时使用，直接调用 loader
type NoopCache struct{}

func (NoopCache) GetOrLoad(ctx context.Context, _ string, _ time.Duration, loader func(context.Context) (any, error), dst any) error {
	val, err := loader(ctx)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	// 尝试直接赋值/解码
	switch typed := dst.(type) {
	case *interface{}:
		*typed = val
	default:
		// 最简单的方式是通过 JSON 再编码一次
		codec := JSONCodec{}
		data, err := codec.Marshal(val)
		if err != nil {
			return err
		}
		return codec.Unmarshal(data, dst)
	}
	return nil
}

func (NoopCache) Delete(context.Context, string)         {}
func (NoopCache) DeleteByPrefix(context.Context, string) {}

// LayeredCache 简单多级缓存 orchestrator（先本地，再远端）
type LayeredCache struct {
	Local      Cache
	Remote     Cache
	Codec      Codec
	DefaultTTL time.Duration
	group      singleflight.Group
}

func NewLayeredCache(local Cache, remote Cache, codec Codec, defaultTTL time.Duration) *LayeredCache {
	if codec == nil {
		codec = JSONCodec{}
	}
	return &LayeredCache{
		Local:      local,
		Remote:     remote,
		Codec:      codec,
		DefaultTTL: defaultTTL,
	}
}

func (c *LayeredCache) GetOrLoad(ctx context.Context, key string, ttl time.Duration, loader func(context.Context) (any, error), dst any) error {
	if c == nil || c.Local == nil {
		return NoopCache{}.GetOrLoad(ctx, key, ttl, loader, dst)
	}

	// 1. local
	if hit, err := c.Local.Get(ctx, key, dst); err == nil && hit {
		return nil
	}

	// 2. remote
	if c.Remote != nil {
		var tmp any
		if hit, err := c.Remote.Get(ctx, key, &tmp); err == nil && hit {
			if dst != nil {
				// decode tmp into dst
				data, err := c.Codec.Marshal(tmp)
				if err != nil {
					return err
				}
				if err := c.Codec.Unmarshal(data, dst); err != nil {
					return err
				}
			}
			_ = c.Local.Set(ctx, key, tmp, ttlOrDefault(ttl, c.DefaultTTL))
			return nil
		}
	}

	// 3. loader with singleflight to avoid thundering herd
	val, err, _ := c.group.Do(key, func() (any, error) {
		return loader(ctx)
	})
	if err != nil {
		return err
	}

	_ = c.Local.Set(ctx, key, val, ttlOrDefault(ttl, c.DefaultTTL))
	if c.Remote != nil {
		_ = c.Remote.Set(ctx, key, val, ttlOrDefault(ttl, c.DefaultTTL))
	}

	if dst != nil {
		data, err := c.Codec.Marshal(val)
		if err != nil {
			return err
		}
		return c.Codec.Unmarshal(data, dst)
	}
	return nil
}

func (c *LayeredCache) Delete(ctx context.Context, key string) {
	if c == nil {
		return
	}
	if c.Local != nil {
		_ = c.Local.Delete(ctx, key)
	}
	if c.Remote != nil {
		_ = c.Remote.Delete(ctx, key)
	}
}

func (c *LayeredCache) DeleteByPrefix(ctx context.Context, prefix string) {
	if c == nil {
		return
	}
	if c.Local != nil {
		_ = c.Local.DeleteByPrefix(ctx, prefix)
	}
	if c.Remote != nil {
		_ = c.Remote.DeleteByPrefix(ctx, prefix)
	}
}

func ttlOrDefault(ttl time.Duration, def time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}
	return def
}
