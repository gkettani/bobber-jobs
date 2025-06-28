package models

import (
	"time"
)

// PipelineStatus represents the overall status of the job processing pipeline
type PipelineStatus struct {
	// Queue information
	QueueSize int `json:"queue_size"`

	// Service capabilities
	DiscoveryCompanies  int `json:"discovery_companies"`
	EnrichmentCompanies int `json:"enrichment_companies"`

	// Runtime information
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`

	// Processing metrics
	Metrics ProcessingMetrics `json:"metrics"`
}

// ProcessingMetrics contains metrics about job processing
type ProcessingMetrics struct {
	// Discovery metrics
	DiscoveryCycles     int64     `json:"discovery_cycles"`
	LastDiscoveryTime   time.Time `json:"last_discovery_time"`
	TotalJobsDiscovered int64     `json:"total_jobs_discovered"`

	// Enrichment metrics
	JobsProcessed  int64 `json:"jobs_processed"`
	JobsSuccessful int64 `json:"jobs_successful"`
	JobsFailed     int64 `json:"jobs_failed"`
	JobsDuplicate  int64 `json:"jobs_duplicate"`

	// Performance metrics
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`

	// Error tracking
	ErrorRate        float64   `json:"error_rate"`
	LastProcessedJob time.Time `json:"last_processed_job"`
}

// ProcessingResult represents the outcome of processing a single job reference
type ProcessingResult struct {
	JobReference   *JobReference    `json:"job_reference"`
	Status         ProcessingStatus `json:"status"`
	JobDetails     *JobDetails      `json:"job_details,omitempty"`
	Error          string           `json:"error,omitempty"`
	ProcessingTime time.Duration    `json:"processing_time"`
	Timestamp      time.Time        `json:"timestamp"`
}

// ProcessingStatus represents the status of job processing
type ProcessingStatus string

const (
	ProcessingStatusSuccess   ProcessingStatus = "success"
	ProcessingStatusFailed    ProcessingStatus = "failed"
	ProcessingStatusDuplicate ProcessingStatus = "duplicate"
	ProcessingStatusSkipped   ProcessingStatus = "skipped"
)

// CalculateSuccessRate calculates the success rate for processing metrics
func (pm *ProcessingMetrics) CalculateSuccessRate() float64 {
	if pm.JobsProcessed == 0 {
		return 0.0
	}
	return float64(pm.JobsSuccessful) / float64(pm.JobsProcessed) * 100.0
}
