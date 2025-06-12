package scraper

import (
	"fmt"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type MistralScraper struct {
	*BaseScraper
}

func NewMistralScraper(baseScraper *BaseScraper) *MistralScraper {
	return &MistralScraper{
		BaseScraper: baseScraper,
	}
}

func (s *MistralScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
	doc, err := s.FetchHTMLWithRetry(jobListing.URL)
	if err != nil {
		return nil, err
	}

	title := doc.Find(".posting-headline").First().Text()
	location := doc.Find(".posting-categories > .location").First().Text()
	description, err := doc.Find(".section-wrapper").Html()
	if err != nil {
		return nil, fmt.Errorf("extracting description: %w", err)
	}

	job := models.Job{
		ExternalID:  jobListing.ExternalID,
		CompanyName: string(common.CompanyNameMistral),
		URL:         jobListing.URL,
		Title:       title,
		Location:    location,
		Description: description,
	}

	return &job, nil

}

func (s *MistralScraper) CanHandle(url string) bool {
	return strings.Contains(url, "mistral")
}

func (s *MistralScraper) CompanyName() common.CompanyName {
	return common.CompanyNameMistral
}
