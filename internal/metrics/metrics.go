package metrics

import (
	"net/http"
	"sync"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	instance *MetricsManager
	once     sync.Once
	mu       sync.Mutex
)

type MetricsManager struct {
	registry     *prometheus.Registry
	metricsStore map[string]any
}

// GetManager returns the singleton instance of MetricsManager
func GetManager() *MetricsManager {
	once.Do(func() {
		instance = &MetricsManager{
			registry:     prometheus.NewRegistry(),
			metricsStore: make(map[string]any),
		}
	})
	return instance
}

func (m *MetricsManager) Initialize() {
	prometheus.DefaultRegisterer = m.registry

	go func() {
		logger.Info("Starting metrics server on port 8080")
		http.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
		http.ListenAndServe(":8080", nil)
	}()
}

func (m *MetricsManager) CreateCounter(name, help string) prometheus.Counter {
	mu.Lock()
	defer mu.Unlock()

	if metric, exists := m.metricsStore[name]; exists {
		return metric.(prometheus.Counter)
	}

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})

	prometheus.MustRegister(counter)

	m.metricsStore[name] = counter
	return counter
}

func (m *MetricsManager) CreateCounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	mu.Lock()
	defer mu.Unlock()

	if metric, exists := m.metricsStore[name]; exists {
		return metric.(*prometheus.CounterVec)
	}

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelNames)

	prometheus.MustRegister(counterVec)

	m.metricsStore[name] = counterVec
	return counterVec
}

func (m *MetricsManager) CreateHistogramVec(name, help string, buckets []float64, labelNames []string) *prometheus.HistogramVec {
	mu.Lock()
	defer mu.Unlock()

	if metric, exists := m.metricsStore[name]; exists {
		return metric.(*prometheus.HistogramVec)
	}

	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelNames)

	prometheus.MustRegister(histogramVec)

	m.metricsStore[name] = histogramVec
	return histogramVec
}

func (m *MetricsManager) CreateGauge(name, help string) prometheus.Gauge {
	mu.Lock()
	defer mu.Unlock()

	if metric, exists := m.metricsStore[name]; exists {
		return metric.(prometheus.Gauge)
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})

	prometheus.MustRegister(gauge)

	m.metricsStore[name] = gauge
	return gauge
}

func (m *MetricsManager) CreateGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec {
	mu.Lock()
	defer mu.Unlock()

	if metric, exists := m.metricsStore[name]; exists {
		return metric.(*prometheus.GaugeVec)
	}

	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labelNames)

	prometheus.MustRegister(gaugeVec)

	m.metricsStore[name] = gaugeVec
	return gaugeVec
}

func init() {
	GetManager().Initialize()
}
