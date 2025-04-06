package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewAirbnbFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameAirbnb,
		"https://careers.airbnb.com/positions-sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/positions/([^<]+)/`)),
	)
}
