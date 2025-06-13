package scraper

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ScraperConfig struct {
	Name        string         `yaml:"name"`
	URLPatterns []string       `yaml:"url_patterns"`
	Selectors   SelectorConfig `yaml:"selectors"`
	Enabled     bool           `yaml:"enabled"`
}

type SelectorConfig struct {
	Title       string `yaml:"title"`
	Location    string `yaml:"location"`
	Description string `yaml:"description"`
}

type ScrapersConfig struct {
	Scrapers map[string]ScraperConfig `yaml:"scrapers"`
}

func LoadScrapersConfig(configPath string) (*ScrapersConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scrapers config file: %w", err)
	}

	var config ScrapersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse scrapers config file: %w", err)
	}

	for key, scraperConfig := range config.Scrapers {
		if err := scraperConfig.Validate(); err != nil {
			return nil, fmt.Errorf("invalid scraper config for %s: %w", key, err)
		}
	}

	return &config, nil
}

func (c *ScraperConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("scraper name is required")
	}

	if len(c.URLPatterns) == 0 {
		return fmt.Errorf("at least one URL pattern is required")
	}

	if c.Selectors.Title == "" {
		return fmt.Errorf("title selector is required")
	}

	if c.Selectors.Location == "" {
		return fmt.Errorf("location selector is required")
	}

	if c.Selectors.Description == "" {
		return fmt.Errorf("description selector is required")
	}

	return nil
}

func (c *ScraperConfig) MatchesURL(url string) bool {
	for _, pattern := range c.URLPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}
