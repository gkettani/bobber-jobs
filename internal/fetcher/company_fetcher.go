package fetcher

import (
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type CompanyFetcher struct {
	strategy      FetchStrategy
	companyName   CompanyName
	sourceURL     string
	extractorFunc ExtractorFunc
}

// NewCompanyFetcher creates a new company fetcher with the specified strategy
func NewCompanyFetcher(
	strategy FetchStrategy,
	companyName CompanyName,
	sourceURL string,
	extractorFunc ExtractorFunc,
) *CompanyFetcher {
	return &CompanyFetcher{
		strategy:      strategy,
		companyName:   companyName,
		sourceURL:     sourceURL,
		extractorFunc: extractorFunc,
	}
}

// Fetch implements the Fetcher interface
func (f *CompanyFetcher) Fetch() ([]*models.JobListing, error) {
	logger.Info(fmt.Sprintf("Fetching job listings from %s", f.companyName))
	return f.strategy.FetchJobs(f.sourceURL, f.extractorFunc)
}

// CompanyName implements the Fetcher interface
func (f *CompanyFetcher) CompanyName() CompanyName {
	return f.companyName
}
