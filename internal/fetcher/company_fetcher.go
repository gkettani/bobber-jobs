package fetcher

import (
	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type CompanyFetcher struct {
	strategy      FetchStrategy
	companyName   common.CompanyName
	sourceURL     string
	extractorFunc ExtractorFunc
}

func NewCompanyFetcher(
	strategy FetchStrategy,
	companyName common.CompanyName,
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

func (f *CompanyFetcher) Fetch() ([]*models.JobListing, error) {
	return f.strategy.FetchJobs(f.sourceURL, f.extractorFunc)
}

func (f *CompanyFetcher) CompanyName() common.CompanyName {
	return f.companyName
}
