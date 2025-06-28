package orchestration

import (
	"context"
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/queue"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

// Config holds the configuration for the orchestrator
type Config struct {
	DiscoveryInterval time.Duration
	ProcessingDelay   time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		DiscoveryInterval: 10 * time.Minute,
		ProcessingDelay:   100 * time.Millisecond,
	}
}

// Orchestrator coordinates the entire job processing pipeline
type Orchestrator struct {
	config               Config
	discoveryService     services.JobDiscoveryService
	enrichmentService    services.JobEnrichmentService
	persistenceService   services.JobPersistenceService
	deduplicationService services.DeduplicationService
	queue                *queue.JobQueue
	stopChan             chan struct{}

	// Pipeline metrics and status
	startTime time.Time
	metrics   models.ProcessingMetrics
}

// NewOrchestrator creates a new pipeline orchestrator
func NewOrchestrator(
	config Config,
	discoveryService services.JobDiscoveryService,
	enrichmentService services.JobEnrichmentService,
	persistenceService services.JobPersistenceService,
	deduplicationService services.DeduplicationService,
) *Orchestrator {
	return &Orchestrator{
		config:               config,
		discoveryService:     discoveryService,
		enrichmentService:    enrichmentService,
		persistenceService:   persistenceService,
		deduplicationService: deduplicationService,
		queue:                queue.NewJobQueue(),
		stopChan:             make(chan struct{}),
		startTime:            time.Now(),
		metrics:              models.ProcessingMetrics{},
	}
}

// Start begins the orchestrated job processing pipeline
func (o *Orchestrator) Start(ctx context.Context) error {
	logger.Info("Starting job processing pipeline")

	go o.runDiscoveryWorker(ctx)

	go o.runEnrichmentWorker(ctx)

	logger.Info("Job processing pipeline started successfully")
	return nil
}

// Stop gracefully shuts down the orchestrator
func (o *Orchestrator) Stop() error {
	logger.Info("Stopping job processing pipeline")
	close(o.stopChan)
	return nil
}

// runDiscoveryWorker runs the job discovery process periodically
func (o *Orchestrator) runDiscoveryWorker(ctx context.Context) {
	ticker := time.NewTicker(o.config.DiscoveryInterval)
	defer ticker.Stop()

	// Run discovery immediately on startup
	o.runDiscovery(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Discovery worker shutting down due to context cancellation")
			return
		case <-o.stopChan:
			logger.Info("Discovery worker shutting down")
			return
		case <-ticker.C:
			o.runDiscovery(ctx)
		}
	}
}

// runDiscovery performs the job discovery process
func (o *Orchestrator) runDiscovery(ctx context.Context) {
	logger.Info("Starting job discovery cycle")
	startTime := time.Now()

	allJobReferences, err := o.discoveryService.DiscoverJobs(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Error during job discovery: %v", err))
		return
	}

	totalJobs := 0
	for companyName, jobReferences := range allJobReferences {
		logger.Info(fmt.Sprintf("Discovered %d job references for %s", len(jobReferences), companyName))

		for _, jobRef := range jobReferences {
			o.queue.Enqueue(jobRef)
			totalJobs++
		}
	}

	// Update discovery metrics
	o.metrics.DiscoveryCycles++
	o.metrics.LastDiscoveryTime = time.Now()
	o.metrics.TotalJobsDiscovered += int64(totalJobs)

	duration := time.Since(startTime)
	logger.Info(fmt.Sprintf("Discovery cycle completed in %v - added %d job references to queue", duration, totalJobs))
}

// runEnrichmentWorker runs the job enrichment process continuously
func (o *Orchestrator) runEnrichmentWorker(ctx context.Context) {
	logger.Info("Starting enrichment worker")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Enrichment worker shutting down due to context cancellation")
			return
		case <-o.stopChan:
			logger.Info("Enrichment worker shutting down")
			return
		default:
			o.processNextJob(ctx)
		}
	}
}

// processNextJob processes the next job reference from the queue
func (o *Orchestrator) processNextJob(ctx context.Context) {
	if o.queue.IsEmpty() {
		time.Sleep(o.config.ProcessingDelay)
		return
	}

	jobRef := o.queue.Dequeue()
	if jobRef == nil {
		return
	}

	// Track processing start
	startTime := time.Now()
	result := o.processJobReference(ctx, jobRef)

	// Update metrics based on result
	o.updateMetrics(result, time.Since(startTime))

	// Log the result
	o.logProcessingResult(result)
}

// processJobReference processes a single job reference and returns the result
func (o *Orchestrator) processJobReference(ctx context.Context, jobRef *models.JobReference) models.ProcessingResult {
	result := models.ProcessingResult{
		JobReference: jobRef,
		Timestamp:    time.Now(),
	}

	// Check for duplicates
	if o.deduplicationService.IsProcessed(jobRef) {
		result.Status = models.ProcessingStatusDuplicate
		logger.Debug(fmt.Sprintf("Job reference already processed: %s", jobRef.ExternalID))
		return result
	}

	// Mark as processed to prevent reprocessing
	o.deduplicationService.MarkAsProcessed(jobRef)

	logger.Debug(fmt.Sprintf("Processing job reference: %s", jobRef.ExternalID))

	// Enrich the job reference
	jobDetails, err := o.enrichmentService.EnrichJobReference(ctx, jobRef)
	if err != nil {
		result.Status = models.ProcessingStatusFailed
		result.Error = err.Error()
		logger.Error(fmt.Sprintf("Error enriching job reference %s: %v", jobRef.ExternalID, err))
		return result
	}

	logger.Debug(fmt.Sprintf("Successfully enriched job: %s", jobDetails.Title))

	// Persist the job details
	err = o.persistenceService.SaveJobDetails(ctx, jobDetails)
	if err != nil {
		result.Status = models.ProcessingStatusFailed
		result.Error = err.Error()
		logger.Error(fmt.Sprintf("Error persisting job details %s: %v", jobDetails.ExternalID, err))
		return result
	}

	result.Status = models.ProcessingStatusSuccess
	result.JobDetails = jobDetails
	logger.Debug(fmt.Sprintf("Successfully persisted job: %s", jobDetails.Title))

	return result
}

// updateMetrics updates the processing metrics based on the result
func (o *Orchestrator) updateMetrics(result models.ProcessingResult, duration time.Duration) {
	result.ProcessingTime = duration
	o.metrics.JobsProcessed++

	switch result.Status {
	case models.ProcessingStatusSuccess:
		o.metrics.JobsSuccessful++
	case models.ProcessingStatusFailed:
		o.metrics.JobsFailed++
	case models.ProcessingStatusDuplicate:
		o.metrics.JobsDuplicate++
	}

	// Update timing metrics
	o.metrics.TotalProcessingTime += duration
	if o.metrics.JobsProcessed > 0 {
		o.metrics.AverageProcessingTime = o.metrics.TotalProcessingTime / time.Duration(o.metrics.JobsProcessed)
	}

	// Update error rate
	o.metrics.ErrorRate = o.metrics.CalculateSuccessRate()
	o.metrics.LastProcessedJob = time.Now()
}

// logProcessingResult logs the processing result appropriately
func (o *Orchestrator) logProcessingResult(result models.ProcessingResult) {
	switch result.Status {
	case models.ProcessingStatusSuccess:
		logger.Debug(fmt.Sprintf("Successfully processed job %s in %v",
			result.JobReference.ExternalID, result.ProcessingTime))
	case models.ProcessingStatusFailed:
		logger.Error(fmt.Sprintf("Failed to process job %s: %s",
			result.JobReference.ExternalID, result.Error))
	case models.ProcessingStatusDuplicate:
		logger.Debug(fmt.Sprintf("Skipped duplicate job %s",
			result.JobReference.ExternalID))
	}
}

// GetQueueSize returns the current size of the job reference queue
func (o *Orchestrator) GetQueueSize() int {
	return o.queue.Size()
}

// GetStatus returns the current status of the orchestrator using structured models
func (o *Orchestrator) GetStatus() models.PipelineStatus {
	uptime := time.Since(o.startTime)

	return models.PipelineStatus{
		QueueSize:           o.GetQueueSize(),
		DiscoveryCompanies:  len(o.discoveryService.GetRegisteredCompanies()),
		EnrichmentCompanies: len(o.enrichmentService.GetSupportedCompanies()),
		StartTime:           o.startTime,
		Uptime:              uptime.String(),
		Metrics:             o.metrics,
	}
}

// GetMetrics returns the current processing metrics
func (o *Orchestrator) GetMetrics() models.ProcessingMetrics {
	return o.metrics
}
