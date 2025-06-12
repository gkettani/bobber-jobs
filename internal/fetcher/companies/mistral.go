package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewMistralFetcher(strategyFactory *fetcher.StrategyFactory) fetcher.Fetcher {
	fetchStrategy := strategyFactory.NewHTMLStrategy(".posting-title")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNameMistral,
		"https://jobs.lever.co/mistral",
		fetcher.RegexExtractor(regexp.MustCompile(`mistral/([a-z0-9-]+)`)),
	)
}
