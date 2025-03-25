package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewDiabolocomFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewHTMLStrategy(baseFetcher, ".posting-title")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		fetcher.CompanyNameDiabolocom,
		"https://jobs.eu.lever.co/diabolocom",
		fetcher.RegexExtractor(regexp.MustCompile(`diabolocom/([a-z0-9-]+)`)),
	)
}
