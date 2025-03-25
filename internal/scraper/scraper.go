package scraper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/models"
)

type Scraper interface {
	Scrape(jobListing *models.JobListing) (*models.Job, error)
	CanHandle(url string) bool
}

type ScraperConfig struct {
	HTTPTimeout time.Duration
	MaxRetries  int
}

// BaseScraper provides common utility methods for concrete scrapers
type BaseScraper struct {
	config     ScraperConfig
	httpClient *http.Client
}

func NewBaseScraper(config ScraperConfig) *BaseScraper {
	return &BaseScraper{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

func (s *BaseScraper) FetchHTMLWithRetry(url string) (*goquery.Document, error) {
	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		resp, err = s.httpClient.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt == s.config.MaxRetries {
			if err != nil {
				return nil, fmt.Errorf("failed to fetch URL after %d attempts: %w", s.config.MaxRetries, err)
			}
			return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
		}

		// Wait before retrying (could implement exponential backoff)
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}

	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}
