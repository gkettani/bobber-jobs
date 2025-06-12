package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewPigmentFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewHTMLStrategy(".posting-title")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNamePigment,
		"https://jobs.lever.co/pigment",
		fetcher.RegexExtractor(regexp.MustCompile(`pigment/([a-z0-9-]+)`)),
	)
}
