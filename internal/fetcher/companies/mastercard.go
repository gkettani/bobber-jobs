package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewMastercardFetcher1(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap1.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}

func NewMastercardFetcher2(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap2.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}

func NewMastercardFetcher3(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMastercard,
		"https://careers.mastercard.com/us/en/sitemap3.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`/job/(R-\d+)/`)),
	)
}
