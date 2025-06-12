package fetcher

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type StrategyFactory struct {
	httpClient *HTTPService
}

func NewStrategyFactory(httpClient *HTTPService) *StrategyFactory {
	return &StrategyFactory{
		httpClient: httpClient,
	}
}

func (f *StrategyFactory) NewSitemapStrategy() *SitemapStrategy {
	return &SitemapStrategy{
		httpService: f.httpClient,
	}
}

func (f *StrategyFactory) NewHTMLStrategy(linkSelector string) *HTMLStrategy {
	return &HTMLStrategy{
		httpService:  f.httpClient,
		linkSelector: linkSelector,
	}
}

/*  Sitemap Strategy */
type SitemapStrategy struct {
	httpService *HTTPService
}

type Sitemap struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func (s *SitemapStrategy) FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error) {
	content, err := s.httpService.FetchContent(sourceURL)
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

/*  HTML Strategy */
type HTMLStrategy struct {
	httpService  *HTTPService
	linkSelector string
}

func (s *HTMLStrategy) FetchJobs(sourceURL string, extractor ExtractorFunc) ([]*models.JobListing, error) {
	content, err := s.httpService.FetchContent(sourceURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
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

type HTTPService struct {
	httpClient *http.Client
}

func NewHTTPService() *HTTPService {
	return &HTTPService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPService) FetchContent(url string) ([]byte, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
