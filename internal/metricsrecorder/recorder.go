package metricsrecorder

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vendemas/imagekit/internal/database"
)

type projectCounters struct {
	requests        int64
	cacheHits       int64
	cacheMisses     int64
	originTransforms int64
	bandwidthBytes  int64
	errors          int64
	responseTimeSum int64
	responseCount   int64
}

type Recorder struct {
	mu       sync.Mutex
	counters map[int]*projectCounters
	repo     *database.Repo
	stopCh   chan struct{}
	stopped  chan struct{}
}

func NewRecorder(repo *database.Repo) *Recorder {
	return &Recorder{
		counters: make(map[int]*projectCounters),
		repo:     repo,
		stopCh:   make(chan struct{}),
		stopped:  make(chan struct{}),
	}
}

func (r *Recorder) RecordRequest(projectID int, isCacheHit bool, bandwidthBytes int64, durationMs int64, hasError bool, hasTransform bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pc, ok := r.counters[projectID]
	if !ok {
		pc = &projectCounters{}
		r.counters[projectID] = pc
	}

	pc.requests++
	if isCacheHit {
		pc.cacheHits++
	} else {
		pc.cacheMisses++
	}
	if hasTransform {
		pc.originTransforms++
	}
	pc.bandwidthBytes += bandwidthBytes
	if hasError {
		pc.errors++
	}
	pc.responseTimeSum += durationMs
	pc.responseCount++
}

func (r *Recorder) Start(ctx context.Context) {
	go r.loop(ctx)
}

func (r *Recorder) Stop() {
	close(r.stopCh)
	<-r.stopped
}

func (r *Recorder) loop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer close(r.stopped)

	for {
		select {
		case <-ctx.Done():
			r.flush(ctx)
			return
		case <-r.stopCh:
			r.flush(context.Background())
			return
		case <-ticker.C:
			r.flush(ctx)
		}
	}
}

func (r *Recorder) flush(ctx context.Context) {
	r.mu.Lock()
	snap := r.counters
	r.counters = make(map[int]*projectCounters)
	r.mu.Unlock()

	if len(snap) == 0 {
		return
	}

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	hour := now.Hour()

	for projectID, pc := range snap {
		if pc.requests == 0 {
			continue
		}

		pm := &database.ProjectMetrics{
			ProjectID:         projectID,
			Date:              date,
			Hour:              hour,
			Requests:          pc.requests,
			CacheHits:         pc.cacheHits,
			CacheMisses:       pc.cacheMisses,
			OriginTransforms:  pc.originTransforms,
			BandwidthBytes:    pc.bandwidthBytes,
			Errors:            pc.errors,
			ResponseTimeSumMs: pc.responseTimeSum,
			ResponseCount:     pc.responseCount,
		}

		if err := r.repo.UpsertProjectMetrics(ctx, pm); err != nil {
			slog.Error("flush metrics", "project_id", projectID, "error", err)
		}
	}

	slog.Debug("metrics flushed", "projects", len(snap))
}
