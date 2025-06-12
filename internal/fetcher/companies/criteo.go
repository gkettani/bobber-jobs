package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewCriteoFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameCriteo,
		"https://careers.criteo.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/jobs/(r\d+)/`)),
	)
}
