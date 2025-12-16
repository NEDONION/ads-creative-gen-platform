package service

import (
	"testing"

	"ads-creative-gen-platform/internal/models"
)

// mockExperimentRepo implements repository.ExperimentRepository for unit tests.
type mockExperimentRepo struct {
	exp           *models.Experiment
	asset         *models.CreativeAsset
	metrics       []models.ExperimentMetric
	updatedFields map[string]interface{}
}

func (m *mockExperimentRepo) ListExperiments(status string, page, pageSize int) ([]models.Experiment, int64, error) {
	return nil, 0, nil
}
func (m *mockExperimentRepo) CreateExperiment(exp *models.Experiment) error         { return nil }
func (m *mockExperimentRepo) CreateVariants([]models.ExperimentVariant) error       { return nil }
func (m *mockExperimentRepo) FindAssetByID(uint) (*models.CreativeAsset, error)     { return nil, nil }
func (m *mockExperimentRepo) FindAssetByUUID(string) (*models.CreativeAsset, error) { return nil, nil }
func (m *mockExperimentRepo) FindAssetWithTaskByID(id uint) (*models.CreativeAsset, error) {
	return m.asset, nil
}
func (m *mockExperimentRepo) GetExperimentByUUID(uuid string) (*models.Experiment, error) {
	return m.exp, nil
}
func (m *mockExperimentRepo) GetExperimentWithVariants(uuid string) (*models.Experiment, error) {
	return m.exp, nil
}
func (m *mockExperimentRepo) UpdateExperimentFields(uuid string, fields map[string]interface{}) error {
	m.updatedFields = fields
	return nil
}
func (m *mockExperimentRepo) GetMetric(expID uint, creativeID uint) (*models.ExperimentMetric, error) {
	for i := range m.metrics {
		if m.metrics[i].ExperimentID == expID && m.metrics[i].CreativeID == creativeID {
			return &m.metrics[i], nil
		}
	}
	return nil, nil
}
func (m *mockExperimentRepo) SaveMetric(metric *models.ExperimentMetric) error {
	m.metrics = append(m.metrics, *metric)
	return nil
}
func (m *mockExperimentRepo) ListMetrics(expID uint) ([]models.ExperimentMetric, error) {
	return m.metrics, nil
}

func TestAssignReturnsVariantAndAsset(t *testing.T) {
	repo := &mockExperimentRepo{
		exp: &models.Experiment{
			UUIDModel: models.UUIDModel{UUID: "exp-1"},
			Status:    models.ExpActive,
			Variants: []models.ExperimentVariant{
				{
					ExperimentID: 1,
					CreativeID:   10,
					BucketStart:  0,
					BucketEnd:    9999,
				},
			},
		},
		asset: &models.CreativeAsset{},
	}

	svc := NewExperimentServiceWithRepo(repo)
	res, err := svc.Assign("exp-1", "user-key")
	if err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	if res == nil || res.Asset == nil || res.Variant.CreativeID != 10 {
		t.Fatalf("unexpected assign result: %#v", res)
	}
}

func TestUpdateStatusSetsTimestamps(t *testing.T) {
	repo := &mockExperimentRepo{}
	svc := NewExperimentServiceWithRepo(repo)

	if err := svc.UpdateStatus("exp-1", models.ExpActive); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if repo.updatedFields["status"] != models.ExpActive {
		t.Fatalf("status not updated, got %#v", repo.updatedFields)
	}
	if _, ok := repo.updatedFields["start_at"]; !ok {
		t.Fatalf("start_at not set for active status")
	}
}

func TestGetMetricsComputesCTR(t *testing.T) {
	repo := &mockExperimentRepo{
		exp: &models.Experiment{UUIDModel: models.UUIDModel{UUID: "exp-1", ID: 1}},
		metrics: []models.ExperimentMetric{
			{ExperimentID: 1, CreativeID: 2, Impressions: 10, Clicks: 5},
		},
	}
	svc := NewExperimentServiceWithRepo(repo)

	dto, err := svc.GetMetrics("exp-1")
	if err != nil {
		t.Fatalf("GetMetrics returned error: %v", err)
	}
	metrics, ok := dto.(*ExperimentMetricsDTO)
	if !ok {
		t.Fatalf("expected ExperimentMetricsDTO, got %T", dto)
	}
	if len(metrics.Variants) != 1 {
		t.Fatalf("unexpected metrics result: %#v", metrics)
	}
	if metrics.Variants[0].CTR != 0.5 {
		t.Fatalf("expected CTR 0.5, got %v", metrics.Variants[0].CTR)
	}
}

func TestRandBucketRange(t *testing.T) {
	for _, key := range []string{"user1", "user2", ""} {
		b := randBucket(key)
		if b < 0 || b > 9999 {
			t.Fatalf("bucket out of range for key %s: %d", key, b)
		}
	}
}
