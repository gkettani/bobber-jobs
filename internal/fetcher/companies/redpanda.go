package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewRedpandaFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewHTMLStrategy(".job-post .cell a")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameRedpanda,
		"https://job-boards.greenhouse.io/redpandadata",
		fetcher.RegexExtractor(regexp.MustCompile(`redpandadata/jobs/([a-z0-9]+)`)),
	)
}
