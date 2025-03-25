package fetcher

import (
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type CompositeFetcher struct {
	fetchers []Fetcher
}

func NewCompositeFetcher(fetchers ...Fetcher) *CompositeFetcher {
	return &CompositeFetcher{fetchers: fetchers}
}

func (f *CompositeFetcher) Fetch(jobsChan chan<- *models.JobListing) error {
	logger.Info("Fetching job listings from all fetchers")
	for _, fetcher := range f.fetchers {
		jobListings, err := fetcher.Fetch()
		if err != nil {
			logger.Error(fmt.Sprintf("Error fetching job listings from %s: %s", fetcher.CompanyName(), err))
			return err
		}
		logger.Info(fmt.Sprintf("Found %d job listings from %s", len(jobListings), fetcher.CompanyName()))
		for _, jobListing := range jobListings {
			jobsChan <- jobListing
		}
	}
	return nil
}

func (f *CompositeFetcher) AddFetcher(fetcher Fetcher) *CompositeFetcher {
	f.fetchers = append(f.fetchers, fetcher)
	return f
}
