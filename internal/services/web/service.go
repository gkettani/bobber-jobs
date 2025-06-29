package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/handlers"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/services/orchestration"
	"github.com/gkettani/bobber-the-swe/internal/services/query"
)

// WebService provides HTTP API and web interface
type WebService interface {
	Start(ctx context.Context) error
	Stop() error
	GetPort() int
}

type webService struct {
	server         *http.Server
	port           int
	jobHandler     *handlers.JobHandler
	companyHandler *handlers.CompanyHandler
	metricsHandler *handlers.MetricsHandler
}

// Config holds web service configuration
type Config struct {
	Port int    `env:"WEB_SERVICE_PORT" envDefault:"8080"`
	Host string `env:"WEB_SERVICE_HOST" envDefault:"localhost"`
}

// LoadConfig loads the web service configuration
func LoadConfig() *Config {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		logger.Error("Failed to parse web service config", "error", err)
		panic(err)
	}
	return config
}

// NewWebService creates a new web service
func NewWebService(orchestrator *orchestration.Orchestrator) WebService {
	config := LoadConfig()

	queryService := query.NewJobQueryService()
	jobHandler := handlers.NewJobHandler(queryService)
	companyHandler := handlers.NewCompanyHandler(queryService)
	metricsHandler := handlers.NewMetricsHandler(queryService, orchestrator)

	ws := &webService{
		port:           config.Port,
		jobHandler:     jobHandler,
		companyHandler: companyHandler,
		metricsHandler: metricsHandler,
	}

	// API routes
	http.HandleFunc("/api/jobs", jobHandler.ListJobs)
	http.HandleFunc("/api/jobs/", ws.handleJobsAPI)
	http.HandleFunc("/api/companies", companyHandler.ListCompanies)
	http.HandleFunc("/api/companies/", ws.handleCompaniesAPI)
	http.HandleFunc("/api/metrics", metricsHandler.GetPipelineMetrics)
	http.HandleFunc("/api/health", metricsHandler.GetHealthStatus)
	http.HandleFunc("/api/dashboard", metricsHandler.GetDashboardData)

	// Web UI routes
	http.HandleFunc("/", ws.serveHome)
	http.HandleFunc("/jobs", ws.serveJobsPage)
	http.HandleFunc("/jobs/", ws.serveJobDetailPage)
	http.HandleFunc("/companies", ws.serveCompaniesPage)

	ws.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler: http.DefaultServeMux,
	}

	return ws
}

// Start starts the web service
func (ws *webService) Start(ctx context.Context) error {
	logger.Info(fmt.Sprintf("Starting web service on port %d", ws.port))

	// Start server in a goroutine
	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Web service failed", "error", err)
		}
	}()

	// Start a goroutine to handle context cancellation
	go func() {
		<-ctx.Done()
		logger.Info("Web service context cancelled, shutting down...")
		ws.Stop()
	}()

	return nil
}

// Stop stops the web service
func (ws *webService) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ws.server.Shutdown(ctx)
}

// GetPort returns the port
func (ws *webService) GetPort() int {
	return ws.port
}

// handleJobsAPI routes job-related API requests
func (ws *webService) handleJobsAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 3 && parts[1] == "jobs" && parts[2] == "search" {
		// /api/jobs/search
		ws.jobHandler.SearchJobs(w, r)
		return
	}

	if len(parts) == 3 && parts[1] == "jobs" {
		// /api/jobs/{id}
		ws.jobHandler.GetJob(w, r)
		return
	}

	http.NotFound(w, r)
}

// handleCompaniesAPI routes company-related API requests
func (ws *webService) handleCompaniesAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 4 && parts[1] == "companies" && parts[3] == "jobs" {
		// /api/companies/{name}/jobs
		ws.companyHandler.GetCompanyJobs(w, r)
		return
	}

	http.NotFound(w, r)
}

// serveHome serves the home page
func (ws *webService) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Bobber jobs - Job Dashboard</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
		.container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; }
		h1 { color: #333; text-align: center; margin-bottom: 30px; }
		.nav { text-align: center; margin-bottom: 30px; }
		.nav a { display: inline-block; padding: 10px 20px; margin: 0 10px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
		.nav a:hover { background: #0056b3; }
		.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
		.stat-card { background: #f8f9fa; padding: 20px; border-radius: 6px; text-align: center; }
		.stat-number { font-size: 24px; font-weight: bold; color: #007bff; }
		.stat-label { color: #666; margin-top: 5px; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Bobber jobs - Job Dashboard</h1>
		<div class="nav">
			<a href="/jobs">Browse Jobs</a>
			<a href="/companies">Company Stats</a>
			<a href="/api/metrics">Pipeline Metrics</a>
			<a href="/api/jobs">API Jobs</a>
			<a href="/api/dashboard">API Dashboard</a>
		</div>
		<div id="stats" class="stats">
			<div class="stat-card">
				<div class="stat-number" id="total-jobs">Loading...</div>
				<div class="stat-label">Total Jobs</div>
			</div>
			<div class="stat-card">
				<div class="stat-number" id="total-companies">Loading...</div>
				<div class="stat-label">Companies</div>
			</div>
			<div class="stat-card">
				<div class="stat-number" id="success-rate">Loading...</div>
				<div class="stat-label">Success Rate</div>
			</div>
		</div>
		
		<h2>Recent Jobs</h2>
		<div id="recent-jobs">Loading...</div>
	</div>
	
	<script>
		// Load dashboard data
		fetch('/api/dashboard')
			.then(response => response.json())
			.then(data => {
				if (data.success) {
					document.getElementById('total-jobs').textContent = data.data.totalJobs.toLocaleString();
					document.getElementById('total-companies').textContent = data.data.companyStats.length;
					document.getElementById('success-rate').textContent = data.data.successRate.toFixed(1) + '%';
					
					// Display recent jobs
					const recentJobsDiv = document.getElementById('recent-jobs');
					const jobs = data.data.recentJobs.slice(0, 5);
					recentJobsDiv.innerHTML = jobs.map(job => 
						'<div style="border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px;">' +
						'<h4><a href="/jobs/' + job.id + '">' + job.title + '</a></h4>' +
						'<p><strong>' + job.companyName + '</strong> - ' + job.location + '</p>' +
						'<p style="color: #666; font-size: 14px;">Added: ' + new Date(job.firstSeenAt).toLocaleDateString() + '</p>' +
						'</div>'
					).join('');
				}
			})
			.catch(error => console.error('Error loading dashboard:', error));
	</script>
</body>
</html>`
	w.Write([]byte(html))
}

// serveJobsPage serves the jobs listing page
func (ws *webService) serveJobsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Jobs - Bobber jobs</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
		.container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; }
		.job-item { border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px; }
		.job-title { font-size: 18px; font-weight: bold; color: #007bff; text-decoration: none; }
		.job-company { color: #666; margin: 5px 0; }
	</style>
</head>
<body>
	<div class="container">
		<h1><a href="/" style="text-decoration: none; color: inherit;">Bobber jobs</a> - Jobs</h1>
		<div id="jobs-list">Loading...</div>
	</div>
	
	<script>
		fetch('/api/jobs')
			.then(response => response.json())
			.then(data => {
				if (data.success) {
					const jobsDiv = document.getElementById('jobs-list');
					jobsDiv.innerHTML = data.data.jobs.map(job => 
						'<div class="job-item">' +
						'<a href="/jobs/' + job.id + '" class="job-title">' + job.title + '</a>' +
						'<div class="job-company"><strong>' + job.companyName + '</strong> - ' + job.location + '</div>' +
						'</div>'
					).join('');
				}
			})
			.catch(error => console.error('Error loading jobs:', error));
	</script>
</body>
</html>`
	w.Write([]byte(html))
}

// serveJobDetailPage serves individual job details
func (ws *webService) serveJobDetailPage(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.NotFound(w, r)
		return
	}

	jobID := pathParts[1]

	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Job Details - Bobber jobs</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
		.container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; }
		.back-link { display: inline-block; margin-bottom: 20px; color: #007bff; text-decoration: none; }
	</style>
</head>
<body>
	<div class="container">
		<a href="/jobs" class="back-link">Back to Jobs</a>
		<div id="job-details">Loading...</div>
	</div>
	
	<script>
		const jobId = '` + jobID + `';
		
		fetch('/api/jobs/' + jobId)
			.then(response => response.json())
			.then(data => {
				if (data.success) {
					const job = data.data;
					document.getElementById('job-details').innerHTML = 
						'<h1>' + job.title + '</h1>' +
						'<p><strong>Company:</strong> ' + job.companyName + '</p>' +
						'<p><strong>Location:</strong> ' + job.location + '</p>' +
						'<p><strong>Description:</strong></p>' +
						'<div>' + job.description.replace(/\n/g, '<br>') + '</div>';
				} else {
					document.getElementById('job-details').innerHTML = '<p>Job not found.</p>';
				}
			})
			.catch(error => {
				document.getElementById('job-details').innerHTML = '<p>Error loading job details.</p>';
			});
	</script>
</body>
</html>`
	w.Write([]byte(html))
}

// serveCompaniesPage serves the companies page
func (ws *webService) serveCompaniesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Companies - Bobber jobs</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
		.container { max-width: 1000px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; }
		.company-item { border: 1px solid #ddd; padding: 20px; margin: 15px 0; border-radius: 4px; }
		.company-name { font-size: 20px; font-weight: bold; color: #007bff; margin-bottom: 10px; }
	</style>
</head>
<body>
	<div class="container">
		<h1><a href="/" style="text-decoration: none; color: inherit;">Bobber the SWE</a> - Companies</h1>
		<div id="companies-list">Loading...</div>
	</div>
	
	<script>
		fetch('/api/companies')
			.then(response => response.json())
			.then(data => {
				if (data.success) {
					const companiesDiv = document.getElementById('companies-list');
					companiesDiv.innerHTML = data.data.map(company => 
						'<div class="company-item">' +
						'<div class="company-name">' + company.companyName + '</div>' +
						'<p>Total Jobs: ' + company.job_count + '</p>' +
						'</div>'
					).join('');
				}
			})
			.catch(error => console.error('Error loading companies:', error));
	</script>
</body>
</html>`
	w.Write([]byte(html))
}
