package fetcher

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type JobFetcher struct {
	httpClient *http.Client
	companies  map[string]CompanyConfig
	metrics    *JobFetcherMetrics
}

type JobFetcherMetrics struct {
	fetchDuration *prometheus.GaugeVec
	fetchTotal    *prometheus.CounterVec
	fetchErrors   *prometheus.CounterVec
	jobsFound     *prometheus.GaugeVec
}

func NewJobFetcher() *JobFetcher {
	metricsManager := metrics.GetManager()

	fetcherMetrics := &JobFetcherMetrics{
		fetchDuration: metricsManager.CreateGaugeVec(
			"fetcher_fetch_duration_seconds",
			"Duration of job listing fetch in seconds",
			[]string{"company", "fetch_type"},
		),
		fetchTotal: metricsManager.CreateCounterVec(
			"fetcher_fetch_total",
			"Total number of fetch operations",
			[]string{"company", "fetch_type"},
		),
		fetchErrors: metricsManager.CreateCounterVec(
			"fetcher_fetch_errors_total",
			"Total number of fetch errors",
			[]string{"company", "fetch_type", "error_type"},
		),
		jobsFound: metricsManager.CreateGaugeVec(
			"fetcher_jobs_found",
			"Number of jobs found per company",
			[]string{"company"},
		),
	}

	return &JobFetcher{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		companies:  make(map[string]CompanyConfig),
		metrics:    fetcherMetrics,
	}
}

func (f *JobFetcher) LoadFromConfig(configPath string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for companyKey, companyConfig := range config.Companies {
		if !companyConfig.Enabled {
			logger.Info(fmt.Sprintf("Company %s is disabled, skipping", companyKey))
			continue
		}

		f.companies[companyKey] = companyConfig
	}

	return nil
}

func (f *JobFetcher) RegisterCompany(config CompanyConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid company config: %w", err)
	}
	f.companies[config.Name] = config
	return nil
}

func (f *JobFetcher) FetchJobs(companyName string) ([]*models.JobListing, error) {
	config, exists := f.companies[companyName]
	if !exists {
		return nil, fmt.Errorf("company %s not registered", companyName)
	}

	start := time.Now()
	defer func() {
		f.metrics.fetchDuration.WithLabelValues(string(companyName), config.FetchType).Set(time.Since(start).Seconds())
	}()

	f.metrics.fetchTotal.WithLabelValues(string(companyName), config.FetchType).Inc()

	var jobs []*models.JobListing
	var err error

	switch config.FetchType {
	case "sitemap":
		jobs, err = f.fetchFromSitemap(config)
	case "html":
		jobs, err = f.fetchFromHTML(config)
	case "api":
		jobs, err = f.fetchFromAPI(config)
	default:
		err = fmt.Errorf("unsupported fetch type: %s", config.FetchType)
	}

	if err != nil {
		f.metrics.fetchErrors.WithLabelValues(string(companyName), config.FetchType, "fetch_error").Inc()
		return nil, err
	}

	f.metrics.jobsFound.WithLabelValues(string(companyName)).Set(float64(len(jobs)))
	return jobs, nil
}

func (f *JobFetcher) FetchAllJobs() (map[string][]*models.JobListing, error) {
	results := make(map[string][]*models.JobListing)

	for companyName := range f.companies {
		jobs, err := f.FetchJobs(companyName)
		if err != nil {
			fmt.Printf("Error fetching jobs for %s: %v\n", companyName, err)
			continue
		}
		results[companyName] = jobs
	}

	return results, nil
}

func (f *JobFetcher) GetRegisteredCompanies() []string {
	companies := make([]string, 0, len(f.companies))
	for name := range f.companies {
		companies = append(companies, name)
	}
	return companies
}

func (f *JobFetcher) fetchFromSitemap(config CompanyConfig) ([]*models.JobListing, error) {
	resp, err := f.httpClient.Get(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap: %w", err)
	}

	var sitemap struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}

	if err := xml.Unmarshal(body, &sitemap); err != nil {
		return nil, fmt.Errorf("failed to parse sitemap: %w", err)
	}

	var jobs []*models.JobListing
	for _, entry := range sitemap.URLs {
		if externalID := f.extractID(entry.Loc, config.GetCompiledPattern()); externalID != "" {
			jobs = append(jobs, &models.JobListing{
				ExternalID: externalID,
				URL:        entry.Loc,
			})
		}
	}

	return jobs, nil
}

func (f *JobFetcher) fetchFromHTML(config CompanyConfig) ([]*models.JobListing, error) {
	resp, err := f.httpClient.Get(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var jobs []*models.JobListing
	doc.Find(config.LinkSelector).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		if externalID := f.extractID(href, config.GetCompiledPattern()); externalID != "" {
			jobs = append(jobs, &models.JobListing{
				ExternalID: externalID,
				URL:        href,
			})
		}
	})

	return jobs, nil
}

func (f *JobFetcher) fetchFromAPI(config CompanyConfig) ([]*models.JobListing, error) {
	var body io.Reader
	if config.RequestBody != "" {
		body = strings.NewReader(config.RequestBody)
	}

	req, err := http.NewRequest(config.Method, config.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	return f.parseAPIResponse(config, respBody)
}

func (f *JobFetcher) parseAPIResponse(config CompanyConfig, respBody []byte) ([]*models.JobListing, error) {
	var data interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Navigate to jobs array using the configured path
	var jobsArray []interface{}

	if config.JobsPath != "" {
		jobsData, err := f.getNestedValue(data, config.JobsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to find jobs at path %s: %w", config.JobsPath, err)
		}
		var ok bool
		jobsArray, ok = jobsData.([]interface{})
		if !ok {
			return nil, fmt.Errorf("jobs data at path %s is not an array", config.JobsPath)
		}
	} else {
		// If no path specified, assume the response is directly an array
		var ok bool
		jobsArray, ok = data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("response is not an array and no jobs_path specified")
		}
	}

	var jobs []*models.JobListing
	idField := config.IDField
	if idField == "" {
		idField = "id" // default field name
	}

	for _, jobData := range jobsArray {
		jobMap, ok := jobData.(map[string]interface{})
		if !ok {
			continue
		}

		idValue, exists := jobMap[idField]
		if !exists {
			continue
		}

		var externalID string
		switch v := idValue.(type) {
		case string:
			externalID = v
		case float64:
			externalID = strconv.Itoa(int(v))
		case int:
			externalID = strconv.Itoa(v)
		default:
			continue
		}

		jobURL := f.generateJobURL(config, externalID)

		jobs = append(jobs, &models.JobListing{
			ExternalID: externalID,
			URL:        jobURL,
		})
	}

	return jobs, nil
}

// getNestedValue navigates through nested maps using dot notation (e.g., "data.jobs")
func (f *JobFetcher) getNestedValue(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var exists bool
			current, exists = v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot navigate to '%s' in non-object", part)
		}
	}

	return current, nil
}

func (f *JobFetcher) generateJobURL(config CompanyConfig, externalID string) string {
	if config.URLTemplate == "" {
		logger.Error(fmt.Sprintf("URL template is required for company %s", config.Name))
		return ""
	}

	url := strings.ReplaceAll(config.URLTemplate, "{id}", externalID)
	return url
}

func (f *JobFetcher) extractID(url string, pattern *regexp.Regexp) string {
	if pattern == nil {
		return ""
	}

	matches := pattern.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}
