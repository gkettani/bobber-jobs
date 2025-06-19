# Bobber the SWE 🎣

Automated job processing pipeline that discovers, scrapes, and stores job postings from company career pages.

## ⚠️ Disclaimer

**Use this tool responsibly.** Do not flood websites with excessive requests or generate traffic that could impact their performance. Respect rate limits, implement delays, and follow robots.txt guidelines. Misuse may result in IP blocking or legal issues.


## 🎯 Purpose

Bobber the SWE automates the process of:
- **Job Discovery**: Finds job listings from company career pages using multiple strategies (sitemaps, HTML parsing, APIs)
- **Job Enrichment**: Scrapes detailed job information including title, location, and description
- **Data Management**: Stores job data with deduplication and change tracking
- **Monitoring**: Provides metrics and health monitoring for the entire pipeline

## 🏗️ Architecture

![bobber-the-swe-architecture](/assets/architecture.png)

### Core Services

- **Discovery Service**: Fetches job references from company career pages
- **Enrichment Service**: Scrapes full job details from individual job URLs  
- **Persistence Service**: Handles database operations with deduplication
- **Deduplication Service**: Prevents processing the same job multiple times
- **Orchestration Service**: Coordinates the entire pipeline with metrics and health monitoring

## 🚀 Getting Started

### Prerequisites

- Go 1.23+ 
- PostgreSQL database (for persistence)
- Redis (optional, for caching)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/gkettani/bobber-jobs.git
   cd bobber-jobs
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Define environment variables in a .env file (recommended)**
   ```bash
   # Database configuration
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=bobber
   export DB_USER=your_user
   export DB_PASSWORD=your_password
   
   # Redis configuration (optional)
   export REDIS_HOST=localhost
   export REDIS_PORT=6379
   ```

4. **Set environment variables**
   ```bash
   source .env
   ```


5. **Build the application**
   ```bash
   go build -o bobber cmd/main/main.go
   ```

### Running the Application

```bash
# Start infra
docker compose -f ./infra/docker-compose.yml up -d

# Start observability infra
docker compose -f ./infra/docker-compose.observability.yml up -d

# Run with default configuration
./bobber

# Or run directly with Go
go run cmd/main/main.go
```

The application will:
1. Load company configurations from `config/companies.yaml`
2. Load scraper configurations from `config/scrapers.yaml`
3. Start the discovery and enrichment pipeline
4. Begin processing jobs every 10 minutes (configurable)

### Configuration

The application uses two main configuration files:

- **`config/companies.yaml`**: Defines how to discover jobs from each company
- **`config/scrapers.yaml`**: Defines how to extract job details from job pages

## 🏢 Adding New Companies

### Step 1: Add Company to Discovery Configuration

Edit `config/companies.yaml` and add your company entry:

```yaml
companies:
  your_company:
    name: "Your Company Name"
    fetch_type: "sitemap"  # or "html" or "api"
    url: "https://careers.yourcompany.com/sitemap.xml"
    id_pattern: "/jobs/(\\d+)/"
    enabled: true
```

### Supported Fetch Types

#### 1. Sitemap (`fetch_type: "sitemap"`)
For companies that provide XML sitemaps of their job listings:

```yaml
your_company:
  name: "Your Company"
  fetch_type: "sitemap"
  url: "https://careers.yourcompany.com/sitemap.xml"
  id_pattern: "/jobs/(\\d+)/"  # Regex to extract job ID from URL
  enabled: true
```

#### 2. HTML Parsing (`fetch_type: "html"`)  
For scraping job links directly from HTML pages:

```yaml
your_company:
  name: "Your Company"
  fetch_type: "html"
  url: "https://jobs.yourcompany.com/careers"
  link_selector: ".job-listing a"  # CSS selector for job links
  id_pattern: "/job/([a-z0-9-]+)"
  enabled: true
```

#### 3. API Integration (`fetch_type: "api"`)
For companies with public APIs:

```yaml
your_company:
  name: "Your Company"
  fetch_type: "api"
  url: "https://api.yourcompany.com/jobs"
  method: "GET"  # or "POST"
  headers:
    Accept: "application/json"
    Authorization: "Bearer your-token"  # if needed
  jobs_path: "data.jobs"  # Path to jobs array in response
  id_field: "id"  # Field containing job ID
  url_template: "https://yourcompany.com/careers/{id}"
  enabled: true
```

For GraphQL APIs:
```yaml
your_company:
  name: "Your Company"
  fetch_type: "api"
  url: "https://api.yourcompany.com/graphql"
  method: "POST"
  headers:
    Content-Type: "application/json"
  request_body: |
    {
      "query": "{ jobs { id title } }"
    }
  jobs_path: "data.jobs"
  id_field: "id"
  url_template: "https://yourcompany.com/job/{id}"
  enabled: true
```

### Step 2: Add Scraper Configuration

Edit `config/scrapers.yaml` to define how to extract job details:

```yaml
scrapers:
  your_company:
    name: "Your Company"
    url_patterns: ["yourcompany.com"]  # URL patterns this scraper handles
    selectors:
      title: "h1.job-title"          # CSS selector for job title
      location: ".job-location"      # CSS selector for location
      description: ".job-description" # CSS selector for description
    enabled: true
```

### Step 3: Test the Configuration

1. **Enable only your new company** for testing:
   ```yaml
   # In companies.yaml, set enabled: false for other companies
   your_company:
     enabled: true
   ```

2. **Run the application** to test:
   ```bash
   go run cmd/main/main.go
   ```

3. **Check the logs** for discovery and enrichment success:
   ```
   INFO: Discovered 15 job references for Your Company
   INFO: Successfully enriched job: Senior Software Engineer
   ```

### Step 4: Troubleshooting

#### Common Issues

**No jobs discovered:**
- Verify the URL is accessible
- Check if the `id_pattern` regex matches actual URLs
- For HTML parsing, verify the `link_selector` CSS selector
- Enable debug logging to see detailed output

**Jobs discovered but enrichment fails:**
- Verify the `url_patterns` in scrapers.yaml matches the job URLs
- Check if CSS selectors in scrapers.yaml are correct
- Some sites may require headers (User-Agent, etc.)

**Rate limiting:**
- Add delays between requests
- Implement proper User-Agent headers
- Consider using proxies for high-volume scraping

#### Debugging Tips

1. **Test selectors manually:**
   ```bash
   curl -s "https://company.com/job/123" | grep -A5 "job-title"
   ```

2. **Validate regex patterns:**
   ```bash
   echo "https://company.com/job/123" | grep -Eo "/job/(\\d+)/"
   ```

3. **Check if site blocks scrapers:**
   ```bash
   curl -I "https://company.com/careers" -H "User-Agent: Mozilla/5.0..."
   ```

## 📊 Monitoring and Metrics

The application provides comprehensive monitoring:

### Built-in Metrics
- Discovery cycles completed
- Jobs processed, successful, failed, duplicates
- Processing time averages
- Queue size and throughput
- Error rates per company

## 🔧 Development

### Project Structure
```
├── cmd/main/           # Application entry point
├── internal/
│   ├── services/       # Core business logic services
│   ├── models/         # Data models and structures  
│   ├── fetcher/        # Job discovery implementations
│   ├── scraper/        # Job enrichment implementations
│   ├── repository/     # Database operations
│   ├── cache/          # Caching layer
│   └── common/         # Shared utilities
├── config/             # Configuration files
└── infra/              # Infrastructure and deployment
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
# Build optimized binary
go build -ldflags="-w -s" -o bobber cmd/main/main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bobber-linux cmd/main/main.go
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Update documentation as needed
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Built with ❤️ by Ghali Kettani. Happy job hunting! 🎣