package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewMastercardFetcher1(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap1.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}

func NewMastercardFetcher2(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap2.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}

func NewMastercardFetcher3(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap3.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}
