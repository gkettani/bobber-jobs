package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsConfig struct {
	Port    int  `env:"METRICS_PORT" envDefault:"8080"`
	Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
}

var (
	instance *MetricsManager
	once     sync.Once
	mu       sync.Mutex
)

type MetricsManager struct {
	registry     *prometheus.Registry
	metricsStore map[string]any
	config       *MetricsConfig
}

func LoadConfig() *MetricsConfig {
	config := &MetricsConfig{}
	if err := env.Parse(config); err != nil {
		logger.Error("Failed to parse metrics config", "error", err)
		panic(err)
	}
	return config
}

// GetManager returns the singleton instance of MetricsManager
func GetManager() *MetricsManager {
	once.Do(func() {
		config := LoadConfig()
		instance = &MetricsManager{
			registry:     prometheus.NewRegistry(),
			metricsStore: make(map[string]any),
			config:       config,
		}
	})
	return instance
}

func (m *MetricsManager) Initialize() {
	prometheus.DefaultRegisterer = m.registry

	if m.config.Enabled {
		go func() {
			logger.Info("Starting metrics server on port %d", m.config.Port)
			http.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
			http.ListenAndServe(fmt.Sprintf(":%d", m.config.Port), nil)
		}()
	}
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
