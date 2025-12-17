package repository

import (
	"context"
	"time"

	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/models"
)

type CachedTraceRepository struct {
	inner TraceRepository
	cache cache.DataCache
	keys  cache.KeyBuilder
	ttl   time.Duration
}

func NewCachedTraceRepository(inner TraceRepository, c cache.DataCache, ttl time.Duration, disable bool) TraceRepository {
	if c == nil || disable {
		c = cache.NoopCache{}
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &CachedTraceRepository{
		inner: inner,
		cache: c,
		keys:  cache.KeyBuilder{},
		ttl:   ttl,
	}
}

func (r *CachedTraceRepository) List(page, pageSize int, status, modelName, traceID, productName string) ([]models.ModelTrace, int64, error) {
	key := r.keys.TraceList(status, modelName, traceID, productName, page, pageSize)
	var payload struct {
		Traces []models.ModelTrace
		Total  int64
	}
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		list, total, err := r.inner.List(page, pageSize, status, modelName, traceID, productName)
		if err != nil {
			return nil, err
		}
		return struct {
			Traces []models.ModelTrace
			Total  int64
		}{Traces: list, Total: total}, nil
	}, &payload); err != nil {
		return nil, 0, err
	}
	return payload.Traces, payload.Total, nil
}

func (r *CachedTraceRepository) Detail(traceID string) (*models.ModelTrace, error) {
	key := r.keys.TraceDetail(traceID)
	var trace models.ModelTrace
	if err := r.cache.GetOrLoad(context.Background(), key, r.ttl, func(context.Context) (any, error) {
		return r.inner.Detail(traceID)
	}, &trace); err != nil {
		return nil, err
	}
	return &trace, nil
}

func (r *CachedTraceRepository) CreateTrace(trace *models.ModelTrace) error {
	if err := r.inner.CreateTrace(trace); err != nil {
		return err
	}
	r.invalidate(trace.TraceID)
	return nil
}

func (r *CachedTraceRepository) UpdateTrace(traceID string, updates map[string]interface{}) error {
	if err := r.inner.UpdateTrace(traceID, updates); err != nil {
		return err
	}
	r.invalidate(traceID)
	return nil
}

func (r *CachedTraceRepository) AddStep(step *models.ModelTraceStep) error {
	if err := r.inner.AddStep(step); err != nil {
		return err
	}
	r.invalidate(step.TraceID)
	return nil
}

func (r *CachedTraceRepository) RecoverStuckRunning(maxAge time.Duration, markMessage string) (int64, error) {
	affected, err := r.inner.RecoverStuckRunning(maxAge, markMessage)
	if affected > 0 {
		r.cache.DeleteByPrefix(context.Background(), "traces:list:")
	}
	return affected, err
}

func (r *CachedTraceRepository) invalidate(traceID string) {
	r.cache.DeleteByPrefix(context.Background(), "traces:list:")
	if traceID != "" {
		r.cache.Delete(context.Background(), r.keys.TraceDetail(traceID))
	}
}
