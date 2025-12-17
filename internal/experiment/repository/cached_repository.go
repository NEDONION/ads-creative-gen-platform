package repository

import (
	"context"
	"time"

	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/models"
)

// CachedExperimentRepository 为实验相关读路径增加缓存，写路径自动失效。
type CachedExperimentRepository struct {
	inner        ExperimentRepository
	cache        cache.DataCache
	keys         cache.KeyBuilder
	ttl          time.Duration
	metricsTTL   time.Duration
	disableCache bool
}

func NewCachedExperimentRepository(inner ExperimentRepository, c cache.DataCache, ttl time.Duration, metricsTTL time.Duration, disable bool) ExperimentRepository {
	if c == nil || disable {
		c = cache.NoopCache{}
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if metricsTTL <= 0 {
		metricsTTL = ttl
	}
	return &CachedExperimentRepository{
		inner:      inner,
		cache:      c,
		keys:       cache.KeyBuilder{},
		ttl:        ttl,
		metricsTTL: metricsTTL,
	}
}

func (r *CachedExperimentRepository) ListExperiments(status string, page, pageSize int) ([]models.Experiment, int64, error) {
	key := r.keys.ExperimentList(status, page, pageSize)
	var payload struct {
		Experiments []models.Experiment
		Total       int64
	}
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		list, total, err := r.inner.ListExperiments(status, page, pageSize)
		if err != nil {
			return nil, err
		}
		return struct {
			Experiments []models.Experiment
			Total       int64
		}{Experiments: list, Total: total}, nil
	}, &payload); err != nil {
		return nil, 0, err
	}
	return payload.Experiments, payload.Total, nil
}

func (r *CachedExperimentRepository) CreateExperiment(exp *models.Experiment) error {
	if err := r.inner.CreateExperiment(exp); err != nil {
		return err
	}
	r.invalidateLists()
	return nil
}

func (r *CachedExperimentRepository) CreateVariants(variants []models.ExperimentVariant) error {
	if err := r.inner.CreateVariants(variants); err != nil {
		return err
	}
	r.invalidateLists()
	return nil
}

func (r *CachedExperimentRepository) FindAssetByID(id uint) (*models.CreativeAsset, error) {
	return r.inner.FindAssetByID(id)
}

func (r *CachedExperimentRepository) FindAssetByUUID(uuid string) (*models.CreativeAsset, error) {
	return r.inner.FindAssetByUUID(uuid)
}

func (r *CachedExperimentRepository) FindAssetWithTaskByID(id uint) (*models.CreativeAsset, error) {
	key := r.keys.Asset(id)
	var asset models.CreativeAsset
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		return r.inner.FindAssetWithTaskByID(id)
	}, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *CachedExperimentRepository) GetExperimentByUUID(uuid string) (*models.Experiment, error) {
	key := r.keys.Experiment(uuid)
	var exp models.Experiment
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		return r.inner.GetExperimentByUUID(uuid)
	}, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *CachedExperimentRepository) GetExperimentWithVariants(uuid string) (*models.Experiment, error) {
	key := r.keys.Experiment(uuid) + ":variants"
	var exp models.Experiment
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		return r.inner.GetExperimentWithVariants(uuid)
	}, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *CachedExperimentRepository) UpdateExperimentFields(uuid string, fields map[string]interface{}) error {
	if err := r.inner.UpdateExperimentFields(uuid, fields); err != nil {
		return err
	}
	r.invalidateExperiment(uuid)
	r.invalidateLists()
	return nil
}

func (r *CachedExperimentRepository) GetMetric(expID uint, creativeID uint) (*models.ExperimentMetric, error) {
	return r.inner.GetMetric(expID, creativeID)
}

func (r *CachedExperimentRepository) SaveMetric(metric *models.ExperimentMetric) error {
	return r.inner.SaveMetric(metric)
}

func (r *CachedExperimentRepository) ListMetrics(expID uint) ([]models.ExperimentMetric, error) {
	// metrics keyed by exp UUID; need it from experiment fetch; to keep simple, bypass cache here
	return r.inner.ListMetrics(expID)
}

func (r *CachedExperimentRepository) invalidateExperiment(uuid string) {
	r.cache.Delete(context.Background(), r.keys.Experiment(uuid))
	r.cache.Delete(context.Background(), r.keys.Experiment(uuid)+":variants")
}

func (r *CachedExperimentRepository) invalidateLists() {
	r.cache.DeleteByPrefix(context.Background(), "explist:")
}

func (r *CachedExperimentRepository) invalidateMetrics(uuid string) {
	if uuid == "" {
		return
	}
	r.cache.Delete(context.Background(), r.keys.ExperimentMetrics(uuid))
}
