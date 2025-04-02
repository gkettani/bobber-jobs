package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
	"github.com/gkettani/bobber-the-swe/internal/fetcher/companies"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/queue"
	"github.com/gkettani/bobber-the-swe/internal/repository"
	"github.com/gkettani/bobber-the-swe/internal/scraper"
)

func main() {
	logger.Info("Start the app")
	scraperDuration := metrics.GetManager().CreateGaugeVec("scraper_scrape_duration_seconds", "Duration of job listing scrape in seconds", []string{"scraper"})

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

	jobsQueue := queue.NewJobQueue()

	go func() {
		for {
			compositeFetcher.Fetch(jobsQueue)
			// sleep for 1 minute
			time.Sleep(1 * time.Minute)
		}
	}()

	// Consumer goroutine
	go func() {
		for {
			// Non-blocking check
			if jobsQueue.IsEmpty() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process jobs
			jobListing := jobsQueue.Dequeue()
			if jobListing == nil {
				continue
			}

			logger.Debug(fmt.Sprintf("Processing job listing: %v", jobListing))
			scrapeStart := time.Now()
			job, err := compositeScraper.Scrape(jobListing)
			if err != nil {
				logger.Error(fmt.Sprintf("Error scraping job listing: %v", err))
				continue
			}
			scraperDuration.WithLabelValues(job.CompanyName).Set(time.Since(scrapeStart).Seconds())

			logger.Debug(fmt.Sprintf("Scraped job: %v", job))

			err = jobRepository.Upsert(context.Background(), job)
			if err != nil {
				logger.Error(fmt.Sprintf("Error upserting job: %v", err))
				continue
			}
		}
	}()

	select {}
}
