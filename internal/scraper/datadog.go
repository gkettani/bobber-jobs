package scraper

import (
	"fmt"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type DatadogScraper struct {
	*BaseScraper
}

func NewDatadogScraper(baseScraper *BaseScraper) *DatadogScraper {
	return &DatadogScraper{
		BaseScraper: baseScraper,
	}
}

func (s *DatadogScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
	doc, err := s.FetchHTMLWithRetry(jobListing.URL)
	if err != nil {
		return nil, err
	}

	title := doc.Find("h2").First().Text()
	location := doc.Find("p").First().Text()
	description, err := doc.Find(".job-description").Html()
	if err != nil {
		return nil, fmt.Errorf("extracting description: %w", err)
	}

	job := models.Job{
		ExternalID:  jobListing.ExternalID,
		CompanyName: string(common.CompanyNameDatadog),
		URL:         jobListing.URL,
		Title:       title,
		Location:    location,
		Description: description,
	}

	return &job, nil

}

func (s *DatadogScraper) CanHandle(url string) bool {
	return strings.Contains(url, "datadoghq.com/")
}

func (s *DatadogScraper) CompanyName() common.CompanyName {
	return common.CompanyNameDatadog
}
