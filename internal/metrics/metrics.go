package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "imagekit_requests_total",
		Help: "Total image requests",
	}, []string{"project", "format", "has_transform"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "imagekit_request_duration_seconds",
		Help:    "Request duration",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"project", "type"})

	TransformDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "imagekit_transform_duration_seconds",
		Help:    "Image transform duration",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1},
	}, []string{"format"})

	StorageGetDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "imagekit_storage_get_duration_seconds",
		Help:    "Storage GET duration",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"provider"})

	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "imagekit_errors_total",
		Help: "Total errors",
	}, []string{"project", "type"})

	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imagekit_cache_hits_total",
		Help: "Cache hits",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imagekit_cache_misses_total",
		Help: "Cache misses",
	})

	CacheEvictions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imagekit_cache_evictions_total",
		Help: "Cache evictions (TTL or manual)",
	})

	CacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "imagekit_cache_entries",
		Help: "Current number of cached entries",
	})

	ActiveProjects = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "imagekit_active_projects",
		Help: "Number of active projects in cache",
	})
)
