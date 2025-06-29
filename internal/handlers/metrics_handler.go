package handlers

import (
	"net/http"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
	"github.com/gkettani/bobber-the-swe/internal/services/orchestration"
)

type MetricsHandler struct {
	queryService services.JobQueryService
	orchestrator *orchestration.Orchestrator
}

func NewMetricsHandler(queryService services.JobQueryService, orchestrator *orchestration.Orchestrator) *MetricsHandler {
	return &MetricsHandler{
		queryService: queryService,
		orchestrator: orchestrator,
	}
}

// GetPipelineMetrics handles GET /api/metrics
func (h *MetricsHandler) GetPipelineMetrics(w http.ResponseWriter, r *http.Request) {
	// Get pipeline status and metrics from orchestrator
	pipelineStatus := h.orchestrator.GetStatus()
	pipelineMetrics := h.orchestrator.GetMetrics()

	// Get company statistics from database
	companyStats, err := h.queryService.GetCompanyStats(r.Context())
	if err != nil {
		logger.Error("Failed to get company stats for metrics", "error", err)
		companyStats = []models.CompanyStats{} // Continue with empty stats
	}

	// Get total job count
	totalJobs, err := h.queryService.GetTotalJobCount(r.Context())
	if err != nil {
		logger.Error("Failed to get total job count", "error", err)
		totalJobs = 0
	}

	// Create combined dashboard status
	dashboardStatus := &models.WebDashboardStatus{
		IsRunning:        true, // If we're responding, we're running
		TotalJobsStored:  totalJobs,
		LastDiscoveryRun: pipelineMetrics.LastDiscoveryTime,
		CompanyStats:     companyStats,
		ProcessingRate:   h.calculateProcessingRate(pipelineMetrics),
	}

	// Combine all metrics
	response := map[string]interface{}{
		"pipeline_status":   pipelineStatus,
		"pipeline_metrics":  pipelineMetrics,
		"dashboard_status":  dashboardStatus,
		"total_jobs_stored": totalJobs,
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetHealthStatus handles GET /api/health
func (h *MetricsHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	// Simple health check
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    h.orchestrator.GetStatus().Uptime,
	}

	// Try to get a simple count to check database connectivity
	_, err := h.queryService.GetTotalJobCount(r.Context())
	if err != nil {
		health["status"] = "unhealthy"
		health["database_error"] = err.Error()
		h.writeJSONResponse(w, http.StatusServiceUnavailable, models.NewErrorResponse[interface{}]("Service unhealthy"))
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(health))
}

// GetDashboardData handles GET /api/dashboard
func (h *MetricsHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	// Get recent jobs for dashboard
	recentJobs, err := h.queryService.GetRecentJobs(r.Context(), 10)
	if err != nil {
		logger.Error("Failed to get recent jobs for dashboard", "error", err)
		recentJobs = []*models.LightJobDetails{}
	}

	// Get company statistics
	companyStats, err := h.queryService.GetCompanyStats(r.Context())
	if err != nil {
		logger.Error("Failed to get company stats for dashboard", "error", err)
		companyStats = []models.CompanyStats{}
	}

	// Get total job count
	totalJobs, err := h.queryService.GetTotalJobCount(r.Context())
	if err != nil {
		logger.Error("Failed to get total job count for dashboard", "error", err)
		totalJobs = 0
	}

	// Get pipeline metrics
	pipelineMetrics := h.orchestrator.GetMetrics()
	pipelineStatus := h.orchestrator.GetStatus()

	dashboardData := map[string]interface{}{
		"totalJobs":       totalJobs,
		"recentJobs":      recentJobs,
		"companyStats":    companyStats,
		"pipelineMetrics": pipelineMetrics,
		"pipelineStatus":  pipelineStatus,
		"successRate":     pipelineMetrics.CalculateSuccessRate(),
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(dashboardData))
}

// calculateProcessingRate calculates the jobs processing rate per minute
func (h *MetricsHandler) calculateProcessingRate(metrics models.ProcessingMetrics) float64 {
	if metrics.TotalProcessingTime == 0 || metrics.JobsProcessed == 0 {
		return 0.0
	}

	// Calculate jobs per minute
	totalMinutes := metrics.TotalProcessingTime.Minutes()
	if totalMinutes == 0 {
		return 0.0
	}

	return float64(metrics.JobsProcessed) / totalMinutes
}

// writeJSONResponse writes a JSON response
func (h *MetricsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	jobHandler := &JobHandler{} // Reuse the JSON response functionality
	jobHandler.writeJSONResponse(w, statusCode, data)
}

// writeErrorResponse writes an error response
func (h *MetricsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, models.NewErrorResponse[interface{}](message))
}
