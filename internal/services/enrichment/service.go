package enrichment

import (
	"context"
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/scraper"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

// service implements JobEnrichmentService using the existing scraper
type service struct {
	scraper *scraper.Scraper
}

// NewJobEnrichmentService creates a new job enrichment service
func NewJobEnrichmentService(configPath string) (services.JobEnrichmentService, error) {
	jobScraper := scraper.NewScraper()
	if err := jobScraper.LoadFromConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to load enrichment configuration: %w", err)
	}

	logger.Info(fmt.Sprintf("Loaded scrapers for %d companies", len(jobScraper.GetRegisteredCompanies())))

	return &service{
		scraper: jobScraper,
	}, nil
}

// EnrichJobReference scrapes full job details from a job reference
func (s *service) EnrichJobReference(ctx context.Context, jobRef *models.JobReference) (*models.JobDetails, error) {
	if !jobRef.IsValid() {
		return nil, fmt.Errorf("invalid job reference: missing required fields")
	}

	// Use the existing scraper directly
	jobDetails, err := s.scraper.Scrape(ctx, jobRef)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich job reference %s: %w", jobRef.ExternalID, err)
	}

	return jobDetails, nil
}

// GetSupportedCompanies returns list of companies supported for enrichment
func (s *service) GetSupportedCompanies() []string {
	return s.scraper.GetRegisteredCompanies()
}
