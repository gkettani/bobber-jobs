package fetcher

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// CompanyConfig holds all the configuration needed to fetch jobs for a company
type CompanyConfig struct {
	Name         string            `yaml:"name"`
	FetchType    string            `yaml:"fetch_type"`
	URL          string            `yaml:"url"`
	IDPattern    string            `yaml:"id_pattern"`
	LinkSelector string            `yaml:"link_selector,omitempty"`
	Method       string            `yaml:"method,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	RequestBody  string            `yaml:"request_body,omitempty"`
	Enabled      bool              `yaml:"enabled,omitempty"`

	// API response configuration
	JobsPath    string `yaml:"jobs_path,omitempty"`    // JSON path to jobs array
	IDField     string `yaml:"id_field,omitempty"`     // Field name for job ID
	URLTemplate string `yaml:"url_template,omitempty"` // Template for job URLs

	// Compiled regex pattern (not serialized)
	compiledPattern *regexp.Regexp `yaml:"-"`
}

type FetcherConfig struct {
	Companies map[string]CompanyConfig `yaml:"companies"`
}

func LoadConfig(configPath string) (*FetcherConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FetcherConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	for key, companyConfig := range config.Companies {
		if companyConfig.IDPattern != "" {
			pattern, err := regexp.Compile(companyConfig.IDPattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern for company %s: %w", key, err)
			}
			companyConfig.compiledPattern = pattern
		}

		if err := companyConfig.Validate(); err != nil {
			return nil, fmt.Errorf("invalid config for company %s: %w", key, err)
		}

		config.Companies[key] = companyConfig
	}

	return &config, nil
}

func (c *CompanyConfig) GetCompiledPattern() *regexp.Regexp {
	return c.compiledPattern
}

func (c *CompanyConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("company name is required")
	}

	if c.URL == "" {
		return fmt.Errorf("company URL is required")
	}

	switch c.FetchType {
	case "sitemap", "html", "api":
	default:
		return fmt.Errorf("invalid fetch type: %s", c.FetchType)
	}

	if c.FetchType == "html" && c.LinkSelector == "" {
		return fmt.Errorf("link_selector is required for HTML fetch type")
	}

	if c.FetchType == "api" {
		if c.Method == "" {
			return fmt.Errorf("method is required for API fetch type")
		}
	}

	return nil
}
