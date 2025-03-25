package fetcher

import (
	"encoding/xml"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

// SitemapStrategy implements the FetchStrategy for XML sitemaps
type SitemapStrategy struct {
	*BaseFetcher
}

type Sitemap struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func NewSitemapStrategy(baseFetcher *BaseFetcher) *SitemapStrategy {
	return &SitemapStrategy{
		BaseFetcher: baseFetcher,
	}
}

func (s *SitemapStrategy) FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error) {
	content, err := s.fetchContent(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching sitemap content: %w", err)
	}

	return s.parseSitemapXML(content, extractor)
}

func (s *SitemapStrategy) parseSitemapXML(data []byte, extractor ExtractorFunc) ([]*models.JobListing, error) {
	var sitemap Sitemap
	err := xml.Unmarshal(data, &sitemap)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap XML: %w", err)
	}

	var jobListings []*models.JobListing
	for _, entry := range sitemap.URLs {
		externalID, err := extractor(entry.Loc)
		if err != nil {
			continue
		}
		jobListings = append(jobListings, &models.JobListing{ExternalID: externalID, URL: entry.Loc})
	}

	return jobListings, nil
}

// HTMLStrategy implements the FetchStrategy for HTML pages
type HTMLStrategy struct {
	*BaseFetcher
	linkSelector string
}

func NewHTMLStrategy(baseFetcher *BaseFetcher, linkSelector string) *HTMLStrategy {
	return &HTMLStrategy{
		BaseFetcher:  baseFetcher,
		linkSelector: linkSelector,
	}
}

func (s *HTMLStrategy) FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error) {
	doc, err := s.fetchHTML(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching HTML content: %w", err)
	}

	var jobListings []*models.JobListing

	doc.Find(s.linkSelector).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		externalID, err := extractor(href)
		if err != nil {
			return
		}

		jobListings = append(jobListings, &models.JobListing{
			ExternalID: externalID,
			URL:        href,
		})
	})

	return jobListings, nil
}
