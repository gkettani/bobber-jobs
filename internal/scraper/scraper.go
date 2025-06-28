package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type Scraper struct {
	httpClient *http.Client
	companies  map[string]ScraperConfig
	metrics    *ScraperMetrics
}

type ScraperMetrics struct {
	scrapeDuration *prometheus.GaugeVec
	scrapeTotal    *prometheus.CounterVec
	scrapeErrors   *prometheus.CounterVec
}

func NewScraper() *Scraper {
	metricsManager := metrics.GetManager()

	scraperMetrics := &ScraperMetrics{
		scrapeDuration: metricsManager.CreateGaugeVec(
			"scraper_scrape_duration_seconds",
			"Duration of job reference scrape in seconds",
			[]string{"company"},
		),
		scrapeTotal: metricsManager.CreateCounterVec(
			"scraper_scrape_total",
			"Total number of scrape operations",
			[]string{"company"},
		),
		scrapeErrors: metricsManager.CreateCounterVec(
			"scraper_scrape_errors_total",
			"Total number of scrape errors",
			[]string{"company", "error_type"},
		),
	}

	return &Scraper{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		companies: make(map[string]ScraperConfig),
		metrics:   scraperMetrics,
	}
}

func (s *Scraper) LoadFromConfig(configPath string) error {
	config, err := LoadScrapersConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load scrapers config: %w", err)
	}

	for companyKey, companyConfig := range config.Scrapers {
		if !companyConfig.Enabled {
			logger.Info(fmt.Sprintf("Scraper for %s is disabled, skipping", companyKey))
			continue
		}

		s.companies[companyKey] = companyConfig
	}

	logger.Info(fmt.Sprintf("Loaded %d scraper configurations", len(s.companies)))
	return nil
}

func (s *Scraper) Scrape(ctx context.Context, jobReference *models.JobReference) (*models.JobDetails, error) {
	companyConfig := s.findCompanyByURL(jobReference.URL)
	if companyConfig == nil {
		return nil, fmt.Errorf("no scraper configuration found for URL: %s", jobReference.URL)
	}

	start := time.Now()
	defer func() {
		s.metrics.scrapeDuration.WithLabelValues(companyConfig.Name).Set(time.Since(start).Seconds())
	}()

	s.metrics.scrapeTotal.WithLabelValues(companyConfig.Name).Inc()

	job, err := s.scrapeWithConfig(ctx, jobReference, companyConfig)
	if err != nil {
		s.metrics.scrapeErrors.WithLabelValues(companyConfig.Name, "scrape_error").Inc()
		return nil, fmt.Errorf("failed to scrape job from %s: %w", companyConfig.Name, err)
	}

	return job, nil
}

func (s *Scraper) GetRegisteredCompanies() []string {
	companies := make([]string, 0, len(s.companies))
	for _, config := range s.companies {
		companies = append(companies, config.Name)
	}
	return companies
}

func (s *Scraper) findCompanyByURL(url string) *ScraperConfig {
	for _, config := range s.companies {
		if config.MatchesURL(url) {
			return &config
		}
	}
	return nil
}

func (s *Scraper) scrapeWithConfig(ctx context.Context, jobReference *models.JobReference, config *ScraperConfig) (*models.JobDetails, error) {
	var lastErr error
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * baseDelay
			logger.Info(fmt.Sprintf("Retrying scrape for %s (attempt %d/%d) after %v", jobReference.URL, attempt+1, maxRetries, delay))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		job, err := s.attemptScrape(ctx, jobReference, config)
		if err == nil {
			return job, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "no scraper configuration") ||
			strings.Contains(err.Error(), "could not extract job title") ||
			strings.Contains(err.Error(), "status code: 404") ||
			strings.Contains(err.Error(), "status code: 403") {
			break
		}

		logger.Info(fmt.Sprintf("Scrape attempt %d failed for %s: %v", attempt+1, jobReference.URL, err))
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (s *Scraper) attemptScrape(ctx context.Context, jobReference *models.JobReference, config *ScraperConfig) (*models.JobDetails, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", jobReference.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	title := strings.TrimSpace(doc.Find(config.Selectors.Title).First().Text())
	location := strings.TrimSpace(doc.Find(config.Selectors.Location).First().Text())

	description, err := doc.Find(config.Selectors.Description).Html()
	if err != nil {
		return nil, fmt.Errorf("failed to extract description: %w", err)
	}
	description = strings.TrimSpace(description)

	if title == "" {
		return nil, fmt.Errorf("could not extract job title using selector: %s", config.Selectors.Title)
	}

	job := &models.JobDetails{
		ExternalID:  jobReference.ExternalID,
		CompanyName: config.Name,
		URL:         jobReference.URL,
		Title:       title,
		Location:    location,
		Description: description,
	}

	return job, nil
}
