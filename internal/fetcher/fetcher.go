package fetcher

import (
	"fmt"
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/models"
)

// Fetcher defines the contract for fetching job listings
type Fetcher interface {
	Fetch() ([]*models.JobListing, error)
	CompanyName() CompanyName
}

// FetchStrategy defines the contract for different fetching strategies
type FetchStrategy interface {
	FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error)
}

type CompanyName string

const (
	CompanyNameCriteo     CompanyName = "Criteo"
	CompanyNameDatadog    CompanyName = "Datadog"
	CompanyNameRedpanda   CompanyName = "Redpanda"
	CompanyNamePigment    CompanyName = "Pigment"
	CompanyNameDiabolocom CompanyName = "Diabolocom"
	CompanyNameYelp       CompanyName = "Yelp"
	CompanyNameEtsy       CompanyName = "Etsy"
	CompanyNameAirbnb     CompanyName = "Airbnb"
	CompanyNameStripe     CompanyName = "Stripe"
	CompanyNameMastercard CompanyName = "Mastercard"
)

// ExtractorFunc is a function type for extracting external IDs from URLs
type ExtractorFunc func(url string) (string, error)

// RegexExtractor creates an ExtractorFunc from a regex pattern
func RegexExtractor(pattern *regexp.Regexp) ExtractorFunc {
	return func(url string) (string, error) {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) < 2 {
			return "", fmt.Errorf("no external_id found in URL: %s", url)
		}
		return matches[1], nil
	}
}
