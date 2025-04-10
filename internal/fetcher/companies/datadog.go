package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewDatadogFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewSitemapStrategy(baseFetcher)

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameDatadog,
		"https://careers.datadoghq.com/sitemap.xml",
		fetcher.RegexExtractor(regexp.MustCompile(`gh_jid=(\d+)`)),
	)
}
