package repository

import (
	"context"
	"time"

	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/shared"
)

// CachedAssetRepository 装饰器：缓存素材列表，写操作时清理。
type CachedAssetRepository struct {
	inner ports.AssetRepository
	cache cache.DataCache
	keys  cache.KeyBuilder
	ttl   time.Duration
}

func NewCachedAssetRepository(inner ports.AssetRepository, c cache.DataCache, ttl time.Duration) ports.AssetRepository {
	if c == nil {
		c = cache.NoopCache{}
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &CachedAssetRepository{
		inner: inner,
		cache: c,
		keys:  cache.KeyBuilder{},
		ttl:   ttl,
	}
}

func (r *CachedAssetRepository) Create(ctx context.Context, asset *models.CreativeAsset) error {
	if err := r.inner.Create(ctx, asset); err != nil {
		return err
	}
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedAssetRepository) List(ctx context.Context, query shared.ListAssetsQuery) ([]models.CreativeAsset, int64, error) {
	key := r.keys.AssetList(query)
	var payload struct {
		Assets []models.CreativeAsset
		Total  int64
	}
	if err := r.cache.GetOrLoad(ctx, key, r.ttl, func(context.Context) (any, error) {
		list, total, err := r.inner.List(ctx, query)
		if err != nil {
			return nil, err
		}
		return struct {
			Assets []models.CreativeAsset
			Total  int64
		}{Assets: list, Total: total}, nil
	}, &payload); err != nil {
		return nil, 0, err
	}
	return payload.Assets, payload.Total, nil
}

func (r *CachedAssetRepository) DeleteByTaskID(ctx context.Context, taskID uint) error {
	if err := r.inner.DeleteByTaskID(ctx, taskID); err != nil {
		return err
	}
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedAssetRepository) invalidateLists(ctx context.Context) {
	r.cache.DeleteByPrefix(ctx, "asset:list:")
}
