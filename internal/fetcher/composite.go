package fetcher

import (
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/queue"
	"github.com/prometheus/client_golang/prometheus"
)

type CompositeFetcherMetrics struct {
	fetchDuration *prometheus.GaugeVec
}

type CompositeFetcher struct {
	fetchers []Fetcher
	*CompositeFetcherMetrics
}

func NewCompositeFetcher(fetchers ...Fetcher) *CompositeFetcher {
	fetchDuration := metrics.GetManager().CreateGaugeVec("fetcher_fetch_duration_seconds", "Duration of job listing fetch in seconds", []string{"fetcher"})

	return &CompositeFetcher{
		fetchers: fetchers,
		CompositeFetcherMetrics: &CompositeFetcherMetrics{
			fetchDuration: fetchDuration,
		},
	}
}

func (f *CompositeFetcher) Fetch(jobsQueue *queue.JobQueue) error {
	start := time.Now()

	defer func() {
		f.fetchDuration.WithLabelValues("composite").Set(time.Since(start).Seconds())
	}()

	logger.Info("Fetching job listings from all fetchers")
	for _, fetcher := range f.fetchers {
		logger.Info(fmt.Sprintf("Fetching job listings from %s", fetcher.CompanyName()))

		fetcherStart := time.Now()

		jobListings, err := fetcher.Fetch()
		if err != nil {
			logger.Error(fmt.Sprintf("Error fetching job listings from %s: %s", fetcher.CompanyName(), err))
			continue
		}

		logger.Info(fmt.Sprintf("Found %d job listings from %s", len(jobListings), fetcher.CompanyName()))
		for _, jobListing := range jobListings {
			jobsQueue.Enqueue(jobListing)
		}

		f.fetchDuration.WithLabelValues(string(fetcher.CompanyName())).Set(time.Since(fetcherStart).Seconds())
	}
	return nil
}

func (f *CompositeFetcher) AddFetcher(fetcher Fetcher) *CompositeFetcher {
	f.fetchers = append(f.fetchers, fetcher)
	return f
}
