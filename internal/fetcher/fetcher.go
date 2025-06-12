package fetcher

import (
	"fmt"
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type Fetcher interface {
	Fetch() ([]*models.JobListing, error)
	CompanyName() common.CompanyName
}

type FetchStrategy interface {
	FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error)
}

// ExtractorFunc is a function type for extracting external IDs from URLs
type ExtractorFunc func(url string) (string, error)

func RegexExtractor(pattern *regexp.Regexp) ExtractorFunc {
	return func(url string) (string, error) {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) < 2 {
			return "", fmt.Errorf("no external_id found in URL: %s", url)
		}
		return matches[1], nil
	}
}
