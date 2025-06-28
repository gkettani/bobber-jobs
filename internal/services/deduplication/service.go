package deduplication

import (
	"github.com/gkettani/bobber-the-swe/internal/cache"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

// service implements DeduplicationService using the existing cache
type service struct {
	cache cache.Manager
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService() services.DeduplicationService {
	return &service{
		cache: cache.NewInMemoryCache(),
	}
}

// IsProcessed checks if a job reference has already been processed
func (s *service) IsProcessed(jobRef *models.JobReference) bool {
	if jobRef == nil || jobRef.ExternalID == "" {
		return false
	}
	return s.cache.Exists(jobRef.ExternalID)
}

// MarkAsProcessed marks a job reference as processed
func (s *service) MarkAsProcessed(jobRef *models.JobReference) {
	if jobRef == nil || jobRef.ExternalID == "" {
		return
	}
	s.cache.Set(jobRef.ExternalID, jobRef.ExternalID)
}
