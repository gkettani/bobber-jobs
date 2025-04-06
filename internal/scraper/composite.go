package scraper

import (
	"fmt"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type CompositeScraperMetrics struct {
	scrapeDuration *prometheus.GaugeVec
}

type CompositeScraper struct {
	scrapers []Scraper
	*CompositeScraperMetrics
}

func NewCompositeScraper(scrapers ...Scraper) *CompositeScraper {
	scrapeDuration := metrics.GetManager().CreateGaugeVec("scraper_scrape_duration_seconds", "Duration of job listing scrape in seconds", []string{"scraper"})

	return &CompositeScraper{
		scrapers: scrapers,
		CompositeScraperMetrics: &CompositeScraperMetrics{
			scrapeDuration: scrapeDuration,
		},
	}
}

func (s *CompositeScraper) Scrape(jobListing *models.JobListing) (*models.Job, error) {
	for _, scraper := range s.scrapers {
		if scraper.CanHandle(jobListing.URL) {
			scraperStart := time.Now()
			defer func() {
				s.scrapeDuration.WithLabelValues(string(scraper.CompanyName())).Set(time.Since(scraperStart).Seconds())
			}()
			return scraper.Scrape(jobListing)
		}
	}
	return nil, fmt.Errorf("no suitable scraper found for url: %s", jobListing.URL)
}

func (s *CompositeScraper) AddScraper(scraper Scraper) *CompositeScraper {
	s.scrapers = append(s.scrapers, scraper)
	return s
}
