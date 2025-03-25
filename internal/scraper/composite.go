package scraper

import (
	"fmt"

	"github.com/gkettani/bobber-the-swe/internal/models"
)

type CompositeScraper struct {
	scrapers []Scraper
}

func NewCompositeScraper(scrapers ...Scraper) *CompositeScraper {
	return &CompositeScraper{scrapers: scrapers}
}

func (s *CompositeScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
	for _, scraper := range s.scrapers {
		if scraper.CanHandle(jobListing.URL) {
			return scraper.Scrape(jobListing)
		}
	}
	return nil, fmt.Errorf("no suitable scraper found for url: %s", jobListing.URL)
}

func (s *CompositeScraper) AddScraper(scraper Scraper) *CompositeScraper {
	s.scrapers = append(s.scrapers, scraper)
	return s
}
