package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewEtsyFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameEtsy,
		"https://careers.etsy.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`jobs/([^<]+)`)),
	)
}
