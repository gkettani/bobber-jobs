package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/cache"
	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/queue"
	"github.com/gkettani/bobber-the-swe/internal/repository"
	"github.com/gkettani/bobber-the-swe/internal/scraper"
)

func main() {
	logger.Info("Starting job fetcher application")

	jobFetcher := fetcher.NewJobFetcher()
	if err := jobFetcher.LoadFromConfig("config/companies.yaml"); err != nil {
		logger.Error(fmt.Sprintf("Failed to load company configuration: %v", err))
		panic(err)
	}

	logger.Info(fmt.Sprintf("Loaded %d companies from configuration", len(jobFetcher.GetRegisteredCompanies())))

	scraper := scraper.NewScraper()
	if err := scraper.LoadFromConfig("config/scrapers.yaml"); err != nil {
		logger.Error(fmt.Sprintf("Failed to load scraper configuration: %v", err))
		panic(err)
	}

	logger.Info(fmt.Sprintf("Loaded scrapers for %d companies", len(scraper.GetRegisteredCompanies())))

	jobRepository := repository.NewJobRepository(db.GetDBClient(), 100)
	jobsQueue := queue.NewJobQueue()
	cache := cache.NewInMemoryCache()

	go func() {
		for {
			logger.Info("Fetching jobs from all companies")
			allJobs, err := jobFetcher.FetchAllJobs()
			if err != nil {
				logger.Error(fmt.Sprintf("Error fetching all jobs: %v", err))
			} else {
				for companyName, jobs := range allJobs {
					logger.Info(fmt.Sprintf("Found %d jobs for %s", len(jobs), companyName))
					for _, job := range jobs {
						jobsQueue.Enqueue(job)
					}
				}
			}

			logger.Info("Fetch cycle completed, sleeping for 10 minutes")
			time.Sleep(10 * time.Minute)
		}
	}()

	go func() {
		for {
			if jobsQueue.IsEmpty() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			jobListing := jobsQueue.Dequeue()
			if jobListing == nil {
				continue
			}

			if cache.Exists(jobListing.ExternalID) {
				logger.Debug(fmt.Sprintf("Job listing already exists in cache: %v", jobListing))
				continue
			}

			cache.Set(jobListing.ExternalID, jobListing.ExternalID)

			logger.Debug(fmt.Sprintf("Processing job listing: %v", jobListing))
			ctx := context.Background()
			job, err := scraper.Scrape(ctx, jobListing)
			if err != nil {
				logger.Error(fmt.Sprintf("Error scraping job listing: %v", err))
				continue
			}

			logger.Debug(fmt.Sprintf("Scraped job: %v", job))

			// Insert job directly to repository
			err = jobRepository.Insert(ctx, job)
			if err != nil {
				logger.Error(fmt.Sprintf("Error inserting job: %v", err))
				continue
			}
			logger.Debug(fmt.Sprintf("Successfully inserted job: %s", job.Title))
		}
	}()

	select {}
}
