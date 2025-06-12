package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/cache"
	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
	"github.com/gkettani/bobber-the-swe/internal/fetcher/companies"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/queue"
	"github.com/gkettani/bobber-the-swe/internal/repository"
	"github.com/gkettani/bobber-the-swe/internal/scraper"
)

func main() {
	logger.Info("Start the app")

	strategyFactory := fetcher.NewStrategyFactory(fetcher.NewHTTPService())

	compositeFetcher := fetcher.NewCompositeFetcher().
		AddFetcher(companies.NewDatadogFetcher(strategyFactory)).
		AddFetcher(companies.NewMistralFetcher(strategyFactory)).
		AddFetcher(companies.NewPigmentFetcher(strategyFactory))

	baseScraper := scraper.NewBaseScraper(scraper.ScraperConfig{
		HTTPTimeout: 10 * time.Second,
		MaxRetries:  3,
	})

	compositeScraper := scraper.NewCompositeScraper().
		AddScraper(scraper.NewDatadogScraper(baseScraper)).
		AddScraper(scraper.NewMistralScraper(baseScraper)).
		AddScraper(scraper.NewPigmentScraper(baseScraper))

	jobRepository := repository.NewJobRepository(db.GetDBClient(), 100)

	// Create a job processor that handles batch processing
	jobProcessor := NewJobProcessor(jobRepository, 50, 5*time.Second)

	jobsQueue := queue.NewJobQueue()

	cache := cache.NewInMemoryCache()

	go func() {
		for {
			compositeFetcher.Fetch(jobsQueue)
			// sleep for 1 minute
			time.Sleep(10 * time.Minute)
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

			if cache.Exists(jobListing.ExternalID) {
				logger.Debug(fmt.Sprintf("Job listing already exists in cache: %v", jobListing))
				continue
			}

			cache.Set(jobListing.ExternalID, jobListing.ExternalID)

			logger.Debug(fmt.Sprintf("Processing job listing: %v", jobListing))
			job, err := compositeScraper.Scrape(jobListing)
			if err != nil {
				logger.Error(fmt.Sprintf("Error scraping job listing: %v", err))
				continue
			}

			logger.Debug(fmt.Sprintf("Scraped job: %v", job))

			jobProcessor.Add(job)
		}
	}()

	select {}
}

// JobProcessor handles batch processing of jobs
type JobProcessor struct {
	repository    repository.JobRepository
	batchSize     int
	flushInterval time.Duration
	jobs          []*models.Job
	mu            sync.Mutex
	timer         *time.Timer
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(repository repository.JobRepository, batchSize int, flushInterval time.Duration) *JobProcessor {
	jp := &JobProcessor{
		repository:    repository,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		jobs:          make([]*models.Job, 0, batchSize),
	}

	// Start timer for periodic flushing
	jp.timer = time.AfterFunc(jp.flushInterval, jp.timerFlush)

	return jp
}

// Add adds a job to the batch and flushes if batch is full
func (jp *JobProcessor) Add(job *models.Job) {
	jp.mu.Lock()
	defer jp.mu.Unlock()

	jp.jobs = append(jp.jobs, job)

	if len(jp.jobs) >= jp.batchSize {
		logger.Debug(fmt.Sprintf("Reached batch size of %d after %f seconds, flushing", jp.batchSize, jp.flushInterval.Seconds()))
		jp.flush()
	}
}

// timerFlush is called when the timer expires
func (jp *JobProcessor) timerFlush() {
	jp.mu.Lock()
	defer jp.mu.Unlock()

	logger.Debug(fmt.Sprintf("Flushing %d jobs", len(jp.jobs)))
	jp.flush()
	jp.timer.Reset(jp.flushInterval)
}

// flush sends all jobs to repository
func (jp *JobProcessor) flush() {
	if len(jp.jobs) == 0 {
		return
	}

	ctx := context.Background()
	err := jp.repository.BulkInsert(ctx, jp.jobs)
	if err != nil {
		logger.Error(fmt.Sprintf("Error batch upserting jobs: %v", err))
	} else {
		logger.Debug(fmt.Sprintf("Batch upserted %d jobs", len(jp.jobs)))
	}

	// Clear the batch
	jp.jobs = make([]*models.Job, 0, jp.batchSize)
}
