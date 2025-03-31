package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
	"github.com/gkettani/bobber-the-swe/internal/fetcher/companies"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/repository"
	"github.com/gkettani/bobber-the-swe/internal/scraper"
)

func main() {
	logger.Info("Start the app")

	baseFetcher := fetcher.NewBaseFetcher()

	compositeFetcher := fetcher.NewCompositeFetcher().
		AddFetcher(companies.NewCriteoFetcher(baseFetcher)).
		AddFetcher(companies.NewDatadogFetcher(baseFetcher))

	baseScraper := scraper.NewBaseScraper(scraper.ScraperConfig{
		HTTPTimeout: 10 * time.Second,
		MaxRetries:  3,
	})

	compositeScraper := scraper.NewCompositeScraper().
		AddScraper(scraper.NewCriteoScraper(baseScraper)).
		AddScraper(scraper.NewDatadogScraper(baseScraper))

	jobRepository := repository.NewJobRepository(db.GetDBClient(), 100)

	jobsChan := make(chan *models.JobListing, 100)

	go func() {
		for {
			compositeFetcher.Fetch(jobsChan)
			// sleep for 1 minute
			time.Sleep(1 * time.Minute)
		}
	}()

	for jobListing := range jobsChan {
		logger.Debug(fmt.Sprintf("Received job listing: %v", jobListing))
		job, err := compositeScraper.Scrape(jobListing)
		if err != nil {
			logger.Error(fmt.Sprintf("Error scraping job listing: %v", err))
			continue
		}
		logger.Debug(fmt.Sprintf("Scraped job: %v", job))

		err = jobRepository.Upsert(context.Background(), job)
		if err != nil {
			logger.Error(fmt.Sprintf("Error upserting job: %v", err))
			continue
		}
	}

	select {}
}
