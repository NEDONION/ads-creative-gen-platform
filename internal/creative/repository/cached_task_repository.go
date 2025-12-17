package repository

import (
	"context"
	"time"

	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/shared"
)

// CachedTaskRepository 装饰器：为任务查询提供缓存，写操作自动失效。
type CachedTaskRepository struct {
	inner ports.TaskRepository
	cache cache.DataCache
	keys  cache.KeyBuilder
	ttl   time.Duration
}

func NewCachedTaskRepository(inner ports.TaskRepository, c cache.DataCache, ttl time.Duration) ports.TaskRepository {
	if c == nil {
		c = cache.NoopCache{}
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &CachedTaskRepository{
		inner: inner,
		cache: c,
		keys:  cache.KeyBuilder{},
		ttl:   ttl,
	}
}

func (r *CachedTaskRepository) Create(ctx context.Context, task *models.CreativeTask) error {
	if err := r.inner.Create(ctx, task); err != nil {
		return err
	}
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedTaskRepository) GetByID(ctx context.Context, id uint) (*models.CreativeTask, error) {
	return r.inner.GetByID(ctx, id)
}

func (r *CachedTaskRepository) GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error) {
	return r.inner.GetByUUID(ctx, uuid)
}

func (r *CachedTaskRepository) GetByUUIDWithAssets(ctx context.Context, uuid string) (*models.CreativeTask, error) {
	key := r.keys.TaskDetail(uuid)
	var task models.CreativeTask
	if err := r.cache.GetOrLoad(ctx, key, r.ttl, func(context.Context) (any, error) {
		return r.inner.GetByUUIDWithAssets(ctx, uuid)
	}, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *CachedTaskRepository) UpdateStatus(ctx context.Context, id uint, status models.TaskStatus, progress int) error {
	if err := r.inner.UpdateStatus(ctx, id, status, progress); err != nil {
		return err
	}
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedTaskRepository) UpdateProgress(ctx context.Context, id uint, progress int) error {
	return r.inner.UpdateProgress(ctx, id, progress)
}

func (r *CachedTaskRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	if err := r.inner.UpdateFields(ctx, id, fields); err != nil {
		return err
	}
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedTaskRepository) List(ctx context.Context, query shared.ListTasksQuery) ([]models.CreativeTask, int64, error) {
	key := r.keys.TaskList(query)
	var payload struct {
		Tasks []models.CreativeTask
		Total int64
	}
	if err := r.cache.GetOrLoad(ctx, key, r.ttl, func(context.Context) (any, error) {
		list, total, err := r.inner.List(ctx, query)
		if err != nil {
			return nil, err
		}
		return struct {
			Tasks []models.CreativeTask
			Total int64
		}{Tasks: list, Total: total}, nil
	}, &payload); err != nil {
		return nil, 0, err
	}
	return payload.Tasks, payload.Total, nil
}

func (r *CachedTaskRepository) Delete(ctx context.Context, task *models.CreativeTask) error {
	if err := r.inner.Delete(ctx, task); err != nil {
		return err
	}
	r.cache.Delete(ctx, r.keys.TaskDetail(task.UUID))
	r.invalidateLists(ctx)
	return nil
}

func (r *CachedTaskRepository) invalidateLists(ctx context.Context) {
	r.cache.DeleteByPrefix(ctx, "task:list:")
}
