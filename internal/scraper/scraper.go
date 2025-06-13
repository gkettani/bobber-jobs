package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings" // je veux bien un moyen de savoir combien il y a de job listing attente de scraper
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
			"Duration of job listing scrape in seconds",
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
		httpClient: &http.Client{Timeout: 10 * time.Second},
		companies:  make(map[string]ScraperConfig),
		metrics:    scraperMetrics,
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

func (s *Scraper) Scrape(ctx context.Context, jobListing *models.JobListing) (*models.Job, error) {
	companyConfig := s.findCompanyByURL(jobListing.URL)
	if companyConfig == nil {
		return nil, fmt.Errorf("no scraper configuration found for URL: %s", jobListing.URL)
	}

	start := time.Now()
	defer func() {
		s.metrics.scrapeDuration.WithLabelValues(companyConfig.Name).Set(time.Since(start).Seconds())
	}()

	s.metrics.scrapeTotal.WithLabelValues(companyConfig.Name).Inc()

	job, err := s.scrapeWithConfig(ctx, jobListing, companyConfig)
	if err != nil {
		s.metrics.scrapeErrors.WithLabelValues(companyConfig.Name, "scrape_error").Inc()
		return nil, fmt.Errorf("failed to scrape job from %s: %w", companyConfig.Name, err)
	}

	return job, nil
}

func (s *Scraper) CanHandle(url string) bool {
	return s.findCompanyByURL(url) != nil
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

func (s *Scraper) scrapeWithConfig(ctx context.Context, jobListing *models.JobListing, config *ScraperConfig) (*models.Job, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", jobListing.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	job := &models.Job{
		ExternalID:  jobListing.ExternalID,
		CompanyName: config.Name,
		URL:         jobListing.URL,
		Title:       title,
		Location:    location,
		Description: description,
	}

	return job, nil
}
