package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewEtsyFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameEtsy,
		"https://careers.etsy.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`jobs/([^<]+)`)),
	)
}
