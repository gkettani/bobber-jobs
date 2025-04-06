package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewYelpFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameYelp,
		"https://www.yelp.careers/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/us/en/job/(\d+)/`)),
	)
}
