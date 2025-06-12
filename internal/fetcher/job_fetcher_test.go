package fetcher

import (
	"testing"
)

func TestJobFetcher_RegisterCompany(t *testing.T) {
	fetcher := NewJobFetcher()

	// Test valid company registration
	config := CompanyConfig{
		Name:      "test-company",
		FetchType: "sitemap",
		URL:       "https://test-company.com/sitemap.xml",
		IDPattern: "/job/(\\d+)",
	}

	err := fetcher.RegisterCompany(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test that company was registered
	companies := fetcher.GetRegisteredCompanies()
	if len(companies) != 1 {
		t.Fatalf("Expected 1 company, got: %d", len(companies))
	}

	if companies[0] != "test-company" {
		t.Fatalf("Expected 'test-company', got: %s", companies[0])
	}
}

func TestJobFetcher_RegisterCompany_Invalid(t *testing.T) {
	fetcher := NewJobFetcher()

	// Test invalid company registration (missing name)
	config := CompanyConfig{
		FetchType: "sitemap",
		URL:       "https://test-company.com/sitemap.xml",
		IDPattern: "/job/(\\d+)",
	}

	err := fetcher.RegisterCompany(config)
	if err == nil {
		t.Fatal("Expected error for invalid config, got nil")
	}
}

func TestJobFetcher_FetchJobs_UnregisteredCompany(t *testing.T) {
	fetcher := NewJobFetcher()

	_, err := fetcher.FetchJobs("nonexistent")
	if err == nil {
		t.Fatal("Expected error for unregistered company, got nil")
	}

	expectedError := "company nonexistent not registered"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestCompanyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CompanyConfig
		wantErr bool
	}{
		{
			name: "valid sitemap config",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "sitemap",
				URL:       "https://test.com/sitemap.xml",
				IDPattern: "/job/(\\d+)",
			},
			wantErr: false,
		},
		{
			name: "valid html config",
			config: CompanyConfig{
				Name:         "test",
				FetchType:    "html",
				URL:          "https://test.com/jobs",
				LinkSelector: ".job-link",
				IDPattern:    "/job/(\\d+)",
			},
			wantErr: false,
		},
		{
			name: "valid api config",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "api",
				URL:       "https://test.com/api",
				Method:    "POST",
				JobsPath:  "data.jobs",
				IDField:   "id",
			},
			wantErr: false,
		},
		{
			name: "valid api config without jobs_path",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "api",
				URL:       "https://test.com/api",
				Method:    "GET",
				IDField:   "id",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: CompanyConfig{
				FetchType: "sitemap",
				URL:       "https://test.com/sitemap.xml",
			},
			wantErr: true,
		},
		{
			name: "missing url",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "sitemap",
			},
			wantErr: true,
		},
		{
			name: "invalid fetch type",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "invalid",
				URL:       "https://test.com",
			},
			wantErr: true,
		},
		{
			name: "html missing link selector",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "html",
				URL:       "https://test.com/jobs",
			},
			wantErr: true,
		},
		{
			name: "api missing method",
			config: CompanyConfig{
				Name:      "test",
				FetchType: "api",
				URL:       "https://test.com/api",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CompanyConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
