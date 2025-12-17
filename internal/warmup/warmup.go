package warmup

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"ads-creative-gen-platform/internal/creative/service"
	expsvc "ads-creative-gen-platform/internal/experiment/service"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/tracing"
	"gorm.io/gorm"
)

// Record 记录一次预热执行情况
type Record struct {
	StartedAt  time.Time `json:"started_at"`
	DurationMs int64     `json:"duration"` // 毫秒
	Success    bool      `json:"success"`
	Errors     []string  `json:"errors,omitempty"`
	ActionsRun []string  `json:"actions_run"`
}

// Stats 预热状态
type Stats struct {
	Runs        int        `json:"runs"`
	Successes   int        `json:"successes"`
	Failures    int        `json:"failures"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	LastSuccess *time.Time `json:"last_success,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
	Recent      []Record   `json:"recent"`
}

// Manager 管理预热任务和状态
type Manager struct {
	cfg    Config
	target Targets

	mu    sync.Mutex
	stats Stats
	limit int
}

// Config 预热配置
type Config struct {
	Interval time.Duration
	Timeout  time.Duration
}

// Targets 需要预热的组件
type Targets struct {
	DB         *sql.DB
	GormDB     *gorm.DB
	Creative   *service.CreativeService
	Experiment *expsvc.ExperimentService
	Trace      *tracing.TraceService
}

func envBool(key string, def bool) bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if val == "true" || val == "1" || val == "yes" {
		return true
	}
	if val == "false" || val == "0" || val == "no" {
		return false
	}
	return def
}

// New 创建 Manager
func New(cfg Config, target Targets) *Manager {
	if cfg.Interval <= 0 {
		cfg.Interval = 3 * time.Minute
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	m := &Manager{
		cfg:    cfg,
		target: target,
		limit:  10,
	}
	m.loadRecentFromDB()
	return m
}

// Start 异步启动预热循环
func (m *Manager) Start() {
	go func() {
		// 立即跑一次
		m.runOnce()
		ticker := time.NewTicker(m.cfg.Interval)
		defer ticker.Stop()
		for range ticker.C {
			m.runOnce()
		}
	}()
}

// RunNow 立即执行一次预热
func (m *Manager) RunNow() {
	m.runOnce()
}

// Stats 返回当前状态
func (m *Manager) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stats
}

func (m *Manager) runOnce() {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.Timeout)
	defer cancel()

	skipDB := envBool("WARMUP_SKIP_DB", false)

	var actions []string
	var errs []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	addErr := func(msg string) {
		mu.Lock()
		errs = append(errs, msg)
		mu.Unlock()
	}
	addAction := func(msg string) {
		mu.Lock()
		actions = append(actions, msg)
		mu.Unlock()
	}

	// 1) DB ping + 核心表轻查询
	if m.target.DB != nil && !skipDB {
		addAction("db:ping")
		pingOK := true
		if _, err := m.target.DB.ExecContext(ctx, "SELECT 1"); err != nil {
			addErr("db ping: " + err.Error())
			pingOK = false
		}

		if pingOK {
			coreTables := []string{"creative_tasks", "creative_assets", "experiments", "model_traces"}
			for _, tbl := range coreTables {
				t := tbl
				wg.Add(1)
				go func() {
					defer wg.Done()
					addAction("db:" + t)
					if _, err := m.target.DB.ExecContext(ctx, "SELECT id FROM "+t+" LIMIT 1"); err != nil {
						addErr(t + ": " + err.Error())
					}
				}()
			}
		} else {
			addErr("skip core table checks due to ping failure")
		}
	}

	// 2) 续命缓存：第一页任务/素材/实验/trace
	if m.target.Creative != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addAction("cache:tasks")
			_, _, err := m.target.Creative.ListAllTasks(service.ListTasksQuery{Page: 1, PageSize: 20})
			if err != nil {
				addErr("tasks: " + err.Error())
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			addAction("cache:assets")
			_, _, err := m.target.Creative.ListAllAssets(service.ListAssetsQuery{Page: 1, PageSize: 20})
			if err != nil {
				addErr("assets: " + err.Error())
			}
		}()
	}

	if m.target.Experiment != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addAction("cache:experiments")
			if _, err := m.target.Experiment.ListExperiments(1, 20, ""); err != nil {
				addErr("experiments: " + err.Error())
			}
		}()
	}

	if m.target.Trace != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addAction("cache:traces")
			if _, err := m.target.Trace.List(1, 20, "", "", "", ""); err != nil {
				addErr("traces: " + err.Error())
			}
		}()
	}

	wg.Wait()

	record := Record{
		StartedAt:  start,
		DurationMs: time.Since(start).Milliseconds(),
		Success:    len(errs) == 0,
		Errors:     errs,
		ActionsRun: actions,
	}
	m.saveRecord(record)
	if record.Success {
		log.Printf("warmup ok in %dms", record.DurationMs)
	} else {
		log.Printf("warmup failed in %dms, errors: %v", record.DurationMs, errs)
	}
}

func (m *Manager) saveRecord(rec Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.Runs++
	if rec.Success {
		m.stats.Successes++
		t := rec.StartedAt
		m.stats.LastSuccess = &t
	} else {
		m.stats.Failures++
		if len(rec.Errors) > 0 {
			m.stats.LastError = rec.Errors[len(rec.Errors)-1]
		}
	}
	t := rec.StartedAt
	m.stats.LastRun = &t

	// append recent records (ring buffer limited by limit)
	m.stats.Recent = append([]Record{rec}, m.stats.Recent...)
	if len(m.stats.Recent) > m.limit {
		m.stats.Recent = m.stats.Recent[:m.limit]
	}

	// 持久化
	if m.target.GormDB != nil {
		actionsJSON, _ := json.Marshal(rec.ActionsRun)
		errorsJSON, _ := json.Marshal(rec.Errors)
		model := models.WarmupRecord{
			StartedAt:  rec.StartedAt,
			DurationMs: rec.DurationMs,
			Success:    rec.Success,
			Actions:    string(actionsJSON),
			Errors:     string(errorsJSON),
			CreatedAt:  time.Now(),
		}
		if err := m.target.GormDB.Create(&model).Error; err != nil {
			log.Printf("warmup persist failed: %v", err)
		}
	}
}

// loadRecentFromDB 尝试加载最近记录填充 stats
func (m *Manager) loadRecentFromDB() {
	if m.target.GormDB == nil {
		return
	}
	var recs []models.WarmupRecord
	if err := m.target.GormDB.Order("started_at desc").Limit(20).Find(&recs).Error; err != nil {
		log.Printf("warmup load history failed: %v", err)
		return
	}
	var records []Record
	successes := 0
	failures := 0
	for _, r := range recs {
		var acts, errs []string
		_ = json.Unmarshal([]byte(r.Actions), &acts)
		_ = json.Unmarshal([]byte(r.Errors), &errs)
		records = append(records, Record{
			StartedAt:  r.StartedAt,
			DurationMs: r.DurationMs,
			Success:    r.Success,
			Errors:     errs,
			ActionsRun: acts,
		})
		if r.Success {
			successes++
		} else {
			failures++
		}
	}
	m.stats.Recent = records
	m.stats.Runs = len(records)
	m.stats.Successes = successes
	m.stats.Failures = failures
	if len(records) > 0 {
		t := records[0].StartedAt
		m.stats.LastRun = &t
		if records[0].Success {
			m.stats.LastSuccess = &t
		} else {
			// 查找最近一次成功
			for _, rec := range records {
				if rec.Success {
					successTime := rec.StartedAt
					m.stats.LastSuccess = &successTime
					break
				}
			}
		}
		if len(records[0].Errors) > 0 {
			m.stats.LastError = records[0].Errors[len(records[0].Errors)-1]
		}
	}
}
