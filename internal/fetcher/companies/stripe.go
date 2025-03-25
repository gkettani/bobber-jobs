package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewStripeFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		fetcher.CompanyNameStripe,
		"https://stripe.com/sitemap/partition-0.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/jobs/listing/[^/]+/(\d+)`)),
	)
}
