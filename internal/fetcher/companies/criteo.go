package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewCriteoFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameCriteo,
		"https://careers.criteo.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/jobs/(r\d+)/`)),
	)
}
