package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewDatadogFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewSitemapStrategy()

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameDatadog,
		"https://careers.datadoghq.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`gh_jid=(\d+)`)),
	)
}
