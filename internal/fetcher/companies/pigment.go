package companies

import (
	"regexp"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/fetcher"
)

func NewPigmentFetcher(baseFetcher *fetcher.BaseFetcher) fetcher.Fetcher {
	fetchStrategy := fetcher.NewHTMLStrategy(baseFetcher, ".posting-title")

	return fetcher.NewCompanyFetcher(
		fetchStrategy,
		common.CompanyNamePigment,
		"https://jobs.lever.co/pigment",
		fetcher.RegexExtractor(regexp.MustCompile(`pigment/([a-z0-9-]+)`)), // Pattern to extract ID
	)
}
