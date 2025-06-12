package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewYelpFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameYelp,
		"https://www.yelp.careers/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/us/en/job/(\d+)/`)),
	)
}
