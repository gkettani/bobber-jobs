package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewAirbnbFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameAirbnb,
		"https://careers.airbnb.com/positions-sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/positions/([^<]+)/`)),
	)
}
