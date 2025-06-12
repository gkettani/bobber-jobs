package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewStripeFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameStripe,
		"https://stripe.com/sitemap/partition-0.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/jobs/listing/[^/]+/(\d+)`)),
	)
}
