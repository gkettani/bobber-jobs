package scraper

import (
	"testing"
)

func TestScraperConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ScraperConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ScraperConfig{
				Name:        "Test Company",
				URLPatterns: []string{"test.com"},
				Selectors: SelectorConfig{
					Title:       "h1",
					Location:    ".location",
					Description: ".description",
				},
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: ScraperConfig{
				URLPatterns: []string{"test.com"},
				Selectors: SelectorConfig{
					Title:       "h1",
					Location:    ".location",
					Description: ".description",
				},
			},
			wantErr: true,
		},
		{
			name: "missing url patterns",
			config: ScraperConfig{
				Name: "Test Company",
				Selectors: SelectorConfig{
					Title:       "h1",
					Location:    ".location",
					Description: ".description",
				},
			},
			wantErr: true,
		},
		{
			name: "missing title selector",
			config: ScraperConfig{
				Name:        "Test Company",
				URLPatterns: []string{"test.com"},
				Selectors: SelectorConfig{
					Location:    ".location",
					Description: ".description",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScraperConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScraperConfig_MatchesURL(t *testing.T) {
	config := ScraperConfig{
		URLPatterns: []string{"example.com", "test.org"},
	}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "matches first pattern",
			url:  "https://jobs.example.com/job/123",
			want: true,
		},
		{
			name: "matches second pattern",
			url:  "https://careers.test.org/position/456",
			want: true,
		},
		{
			name: "no match",
			url:  "https://jobs.other.com/job/789",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.MatchesURL(tt.url); got != tt.want {
				t.Errorf("ScraperConfig.MatchesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScraper_CanHandle(t *testing.T) {
	scraper := NewScraper()

	// Add test configuration
	scraper.companies["test"] = ScraperConfig{
		Name:        "Test Company",
		URLPatterns: []string{"test.com"},
		Selectors: SelectorConfig{
			Title:       "h1",
			Location:    ".location",
			Description: ".description",
		},
		Enabled: true,
	}

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "can handle matching URL",
			url:  "https://jobs.test.com/job/123",
			want: true,
		},
		{
			name: "cannot handle non-matching URL",
			url:  "https://jobs.other.com/job/123",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scraper.CanHandle(tt.url); got != tt.want {
				t.Errorf("UniversalScraper.CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScraper_GetRegisteredCompanies(t *testing.T) {
	scraper := NewScraper()

	// Add test configurations
	scraper.companies["test1"] = ScraperConfig{Name: "Test Company 1"}
	scraper.companies["test2"] = ScraperConfig{Name: "Test Company 2"}

	companies := scraper.GetRegisteredCompanies()

	if len(companies) != 2 {
		t.Errorf("Expected 2 companies, got %d", len(companies))
	}

	// Check that both company names are present
	companyMap := make(map[string]bool)
	for _, company := range companies {
		companyMap[company] = true
	}

	if !companyMap["Test Company 1"] {
		t.Error("Expected 'Test Company 1' to be in registered companies")
	}

	if !companyMap["Test Company 2"] {
		t.Error("Expected 'Test Company 2' to be in registered companies")
	}
}
