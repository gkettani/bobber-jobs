package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewRedpandaFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewHTMLStrategy(baseFetcher, ".job-post .cell a")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		fetcher.CompanyNameRedpanda,
		"https://job-boards.greenhouse.io/redpandadata",
		fetcher.RegexExtractor(regexp.MustCompile(`redpandadata/jobs/([a-z0-9]+)`)), // Pattern to extract ID
	)
}
