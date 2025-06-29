package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

type JobHandler struct {
	queryService services.JobQueryService
}

func NewJobHandler(queryService services.JobQueryService) *JobHandler {
	return &JobHandler{
		queryService: queryService,
	}
}

// ListJobs handles GET /api/jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := h.parseJobFilters(r.URL.Query())
	pagination := h.parsePagination(r.URL.Query())

	// Get jobs from database
	jobList, err := h.queryService.GetJobs(r.Context(), filters, pagination)
	if err != nil {
		logger.Error("Failed to get jobs", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve jobs")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(jobList))
}

// GetJob handles GET /api/jobs/{id}
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	idStr := pathParts[2] // /api/jobs/{id}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	// Get job from database
	job, err := h.queryService.GetJobByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "Job not found")
			return
		}
		logger.Error("Failed to get job", "error", err, "job_id", id)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve job")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(job))
}

// SearchJobs handles GET /api/jobs/search
func (h *JobHandler) SearchJobs(w http.ResponseWriter, r *http.Request) {
	// Get search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Search query is required")
		return
	}

	pagination := h.parsePagination(r.URL.Query())

	// Perform search
	jobList, err := h.queryService.SearchJobs(r.Context(), searchQuery, pagination)
	if err != nil {
		logger.Error("Failed to search jobs", "error", err, "query", searchQuery)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search jobs")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(jobList))
}

// parseJobFilters parses job filters from query parameters
func (h *JobHandler) parseJobFilters(params url.Values) *models.JobFilters {
	return &models.JobFilters{
		CompanyName: params.Get("company"),
		Location:    params.Get("location"),
		Title:       params.Get("title"),
		DateFrom:    params.Get("date_from"),
		DateTo:      params.Get("date_to"),
		Search:      params.Get("q"),
	}
}

// parsePagination parses pagination parameters from query parameters
func (h *JobHandler) parsePagination(params url.Values) *models.Pagination {
	page, _ := strconv.Atoi(params.Get("page"))
	pageSize, _ := strconv.Atoi(params.Get("page_size"))

	return models.NewPagination(page, pageSize)
}

// writeJSONResponse writes a JSON response
func (h *JobHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}

// writeErrorResponse writes an error response
func (h *JobHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, models.NewErrorResponse[interface{}](message))
}
