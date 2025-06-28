package discovery

import (
	"context"
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/fetcher"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

// service implements JobDiscoveryService using the existing fetcher
type service struct {
	fetcher *fetcher.JobFetcher
}

func NewJobDiscoveryService(configPath string) (services.JobDiscoveryService, error) {
	jobFetcher := fetcher.NewJobFetcher()
	if err := jobFetcher.LoadFromConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to load discovery configuration: %w", err)
	}

	logger.Info(fmt.Sprintf("Loaded %d companies for job discovery", len(jobFetcher.GetRegisteredCompanies())))

	return &service{
		fetcher: jobFetcher,
	}, nil
}

// DiscoverJobs discovers job references from all configured companies
func (s *service) DiscoverJobs(ctx context.Context) (map[string][]*models.JobReference, error) {
	allJobListings, err := s.fetcher.FetchAllJobs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job listings: %w", err)
	}

	result := make(map[string][]*models.JobReference)
	for companyName, jobListings := range allJobListings {
		jobReferences := make([]*models.JobReference, 0, len(jobListings))
		for _, jobListing := range jobListings {
			jobRef := &models.JobReference{
				URL:         jobListing.URL,
				ExternalID:  jobListing.ExternalID,
				CompanyName: companyName,
			}
			jobReferences = append(jobReferences, jobRef)
		}
		result[companyName] = jobReferences
	}

	return result, nil
}

// DiscoverJobsForCompany discovers job references for a specific company
func (s *service) DiscoverJobsForCompany(ctx context.Context, companyName string) ([]*models.JobReference, error) {
	jobListings, err := s.fetcher.FetchJobs(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs for company %s: %w", companyName, err)
	}

	jobReferences := make([]*models.JobReference, 0, len(jobListings))
	for _, jobListing := range jobListings {
		jobRef := &models.JobReference{
			URL:         jobListing.URL,
			ExternalID:  jobListing.ExternalID,
			CompanyName: companyName,
		}
		jobReferences = append(jobReferences, jobRef)
	}

	return jobReferences, nil
}

// GetRegisteredCompanies returns list of companies available for discovery
func (s *service) GetRegisteredCompanies() []string {
	return s.fetcher.GetRegisteredCompanies()
}
