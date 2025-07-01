package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/handlers"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/middlewares"
	"github.com/gkettani/bobber-the-swe/internal/services/orchestration"
	"github.com/gkettani/bobber-the-swe/internal/services/query"
)

// WebService provides HTTP API and web interface
type WebService interface {
	Start(ctx context.Context) error
	Stop() error
	GetPort() int
	GetHost() string
}

type webService struct {
	server         *http.Server
	port           int
	host           string
	jobHandler     *handlers.JobHandler
	companyHandler *handlers.CompanyHandler
	metricsHandler *handlers.MetricsHandler
	templates      *template.Template
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

	// Load templates
	templates, err := loadTemplates()
	if err != nil {
		logger.Error("Failed to load templates", "error", err)
		panic(err)
	}

	ws := &webService{
		host:           config.Host,
		port:           config.Port,
		jobHandler:     jobHandler,
		companyHandler: companyHandler,
		metricsHandler: metricsHandler,
		templates:      templates,
	}

	// Setup HTTP routes
	ws.setupRoutes()

	ws.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler: http.DefaultServeMux,
	}

	return ws
}

func (ws *webService) setupRoutes() {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	mux.HandleFunc("/api/jobs", middlewares.WrapHandler(ws.jobHandler.ListJobs))
	mux.HandleFunc("/api/jobs/", middlewares.WrapHandler(ws.handleJobsAPI))
	mux.HandleFunc("/api/companies", middlewares.WrapHandler(ws.companyHandler.ListCompanies))
	mux.HandleFunc("/api/companies/", middlewares.WrapHandler(ws.handleCompaniesAPI))
	mux.HandleFunc("/api/metrics", middlewares.WrapHandler(ws.metricsHandler.GetPipelineMetrics))
	mux.HandleFunc("/api/health", middlewares.WrapHandler(ws.metricsHandler.GetHealthStatus))
	mux.HandleFunc("/api/dashboard", middlewares.WrapHandler(ws.metricsHandler.GetDashboardData))

	mux.HandleFunc("/", middlewares.WrapHandler(ws.serveHome))
	mux.HandleFunc("/jobs", middlewares.WrapHandler(ws.serveJobsPage))
	mux.HandleFunc("/jobs/", middlewares.WrapHandler(ws.serveJobDetailPage))

	// Set the custom mux as the default handler
	http.Handle("/", mux)
}

func loadTemplates() (*template.Template, error) {
	// Check if templates directory exists
	if _, err := os.Stat("web/templates"); os.IsNotExist(err) {
		return nil, fmt.Errorf("templates directory not found: web/templates")
	}

	// Load all templates
	tmpl := template.New("")
	err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(filepath.Base(path), ".html")
			_, err = tmpl.New(name).Parse(string(content))
			if err != nil {
				return fmt.Errorf("error parsing template %s: %w", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}

	return tmpl, nil
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

// GetHost returns the host
func (ws *webService) GetHost() string {
	return ws.host
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

	err := ws.templates.ExecuteTemplate(w, "home", nil)
	if err != nil {
		logger.Error("Error executing home template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// serveJobsPage serves the jobs listing page
func (ws *webService) serveJobsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	err := ws.templates.ExecuteTemplate(w, "jobs", nil)
	if err != nil {
		logger.Error("Error executing jobs template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// serveJobDetailPage serves individual job details
func (ws *webService) serveJobDetailPage(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	err := ws.templates.ExecuteTemplate(w, "job-detail", nil)
	if err != nil {
		logger.Error("Error executing job-detail template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
