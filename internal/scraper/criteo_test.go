package scraper

import (
	"testing"

	"github.com/gkettani/bobber-the-swe/internal/models"
)

func TestCriteoScraper_Scrape(t *testing.T) {
	scraper := NewCriteoScraper(NewBaseScraper(ScraperConfig{}))
	// todo: Properly mock the job listing page to ensure consistent results
	jobListing := &models.JobListing{
		URL: "https://careers.criteo.com/en/jobs/r18353/",
	}

	job, err := scraper.Scrape(jobListing)
	if err != nil {
		t.Errorf("error scraping job: %v", err)
	}

	if job.CompanyName != "Criteo" {
		t.Errorf("company name should be Criteo")
	}

	if job.Title != "Senior Fullstack Software Engineer" {
		t.Errorf("title should be Software Engineer")
	}
}

func TestCriteoScraper_CanHandle(t *testing.T) {
	scraper := NewCriteoScraper(NewBaseScraper(ScraperConfig{}))
	if !scraper.CanHandle("https://careers.criteo.com/en/jobs/r18353/") {
		t.Errorf("scraper should handle this URL")
	}
}
