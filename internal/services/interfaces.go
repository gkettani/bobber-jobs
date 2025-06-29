package services

import (
	"context"

	"github.com/gkettani/bobber-the-swe/internal/models"
)

// JobDiscoveryService finds job references from company career pages
type JobDiscoveryService interface {
	// DiscoverJobs discovers job references from all configured companies
	DiscoverJobs(ctx context.Context) (map[string][]*models.JobReference, error)

	// DiscoverJobsForCompany discovers job references for a specific company
	DiscoverJobsForCompany(ctx context.Context, companyName string) ([]*models.JobReference, error)

	// GetRegisteredCompanies returns list of companies available for discovery
	GetRegisteredCompanies() []string
}

// JobEnrichmentService enriches job references with full details
type JobEnrichmentService interface {
	// EnrichJobReference scrapes full job details from a job reference
	EnrichJobReference(ctx context.Context, jobRef *models.JobReference) (*models.JobDetails, error)

	// GetSupportedCompanies returns list of companies supported for enrichment
	GetSupportedCompanies() []string
}

// JobPersistenceService handles job data persistence
type JobPersistenceService interface {
	// SaveJobDetails saves job details to persistent storage
	SaveJobDetails(ctx context.Context, jobDetails *models.JobDetails) error

	// SaveJobDetailsBatch saves multiple job details in a single transaction
	SaveJobDetailsBatch(ctx context.Context, jobDetails []*models.JobDetails) error
}

// DeduplicationService handles duplicate detection
type DeduplicationService interface {
	// IsProcessed checks if a job reference has already been processed
	IsProcessed(jobRef *models.JobReference) bool

	// MarkAsProcessed marks a job reference as processed
	MarkAsProcessed(jobRef *models.JobReference)
}

// JobQueryService provides read-optimized database operations for the web interface
type JobQueryService interface {
	// GetJobs retrieves jobs with filtering and pagination
	GetJobs(ctx context.Context, filters *models.JobFilters, pagination *models.Pagination) (*models.JobList, error)

	// GetJobByID retrieves a specific job by ID
	GetJobByID(ctx context.Context, id int64) (*models.JobDetails, error)

	// SearchJobs performs full-text search on jobs
	SearchJobs(ctx context.Context, query string, pagination *models.Pagination) (*models.JobList, error)

	// GetCompanyStats retrieves statistics for all companies
	GetCompanyStats(ctx context.Context) ([]models.CompanyStats, error)

	// GetJobCountByCompany returns job counts grouped by company
	GetJobCountByCompany(ctx context.Context) (map[string]int, error)

	// GetTotalJobCount returns the total number of jobs in the database
	GetTotalJobCount(ctx context.Context) (int64, error)

	// GetRecentJobs returns the most recently discovered jobs
	GetRecentJobs(ctx context.Context, limit int) ([]*models.LightJobDetails, error)
}
