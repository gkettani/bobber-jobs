package persistence

import (
	"context"
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/repository"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

// service implements JobPersistenceService using the existing repository
type service struct {
	repository repository.JobRepository
}

// NewJobPersistenceService creates a new job persistence service
func NewJobPersistenceService(batchSize int) services.JobPersistenceService {
	jobRepository := repository.NewJobRepository(db.GetDBClient(), batchSize)

	return &service{
		repository: jobRepository,
	}
}

// SaveJobDetails saves job details to persistent storage
func (s *service) SaveJobDetails(ctx context.Context, jobDetails *models.JobDetails) error {
	if !jobDetails.IsValid() {
		return fmt.Errorf("invalid job details: missing required fields")
	}

	// Use upsert to handle duplicates gracefully
	return s.repository.Upsert(ctx, jobDetails)
}

// SaveJobDetailsBatch saves multiple job details in a single transaction
func (s *service) SaveJobDetailsBatch(ctx context.Context, jobDetailsList []*models.JobDetails) error {
	if len(jobDetailsList) == 0 {
		return nil
	}

	// Validate all job details
	for _, jobDetails := range jobDetailsList {
		if !jobDetails.IsValid() {
			return fmt.Errorf("invalid job details in batch: missing required fields for job %s", jobDetails.ExternalID)
		}
	}

	return s.repository.BulkInsert(ctx, jobDetailsList)
}
