package scraper

import (
	"fmt"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/common"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type PigmentScraper struct {
	*BaseScraper
}

func NewPigmentScraper(baseScraper *BaseScraper) *PigmentScraper {
	return &PigmentScraper{
		BaseScraper: baseScraper,
	}
}

func (s *PigmentScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
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
		CompanyName: string(common.CompanyNamePigment),
		URL:         jobListing.URL,
		Title:       title,
		Location:    location,
		Description: description,
	}

	return &job, nil

}

func (s *PigmentScraper) CanHandle(url string) bool {
	return strings.Contains(url, "pigment")
}

func (s *PigmentScraper) CompanyName() common.CompanyName {
	return common.CompanyNamePigment
}
