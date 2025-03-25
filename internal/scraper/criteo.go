package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type CriteoScraper struct {
	*BaseScraper
}

func NewCriteoScraper(baseScraper *BaseScraper) *CriteoScraper {
	return &CriteoScraper{
		BaseScraper: baseScraper,
	}
}

func (s *CriteoScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
	doc, err := s.FetchHTMLWithRetry(jobListing.URL)
	if err != nil {
		return nil, err
	}

	title := doc.Find("h1").First().Text()
	var locations []string
	doc.Find(".job-meta-locations > li").Each(func(i int, s *goquery.Selection) {
		locations = append(locations, strings.TrimSpace(s.Text()))
	})
	location := strings.Join(locations, "; ")
	description, err := doc.Find(".cms-content").Html()
	if err != nil {
		return nil, fmt.Errorf("extracting description: %w", err)
	}

	job := models.Job{
		ExternalID:  jobListing.ExternalID,
		CompanyName: "Criteo",
		URL:         jobListing.URL,
		Title:       title,
		Location:    location,
		Description: strings.TrimSpace(description),
	}

	return &job, nil

}

func (s *CriteoScraper) CanHandle(url string) bool {
	return strings.Contains(url, "criteo.com")
}
