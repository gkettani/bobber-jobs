package handlers

import (
	"net/http"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
)

type CompanyHandler struct {
	queryService services.JobQueryService
}

func NewCompanyHandler(queryService services.JobQueryService) *CompanyHandler {
	return &CompanyHandler{
		queryService: queryService,
	}
}

// ListCompanies handles GET /api/companies
func (h *CompanyHandler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	stats, err := h.queryService.GetCompanyStats(r.Context())
	if err != nil {
		logger.Error("Failed to get company stats", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve company statistics")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(stats))
}

// GetCompanyJobs handles GET /api/companies/{name}/jobs
func (h *CompanyHandler) GetCompanyJobs(w http.ResponseWriter, r *http.Request) {
	// Extract company name from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid company name")
		return
	}

	companyName := pathParts[2] // /api/companies/{name}/jobs

	// Create filters for this company
	filters := &models.JobFilters{
		CompanyName: companyName,
	}

	// Parse pagination
	jobHandler := &JobHandler{queryService: h.queryService}
	pagination := jobHandler.parsePagination(r.URL.Query())

	// Get jobs for this company
	jobList, err := h.queryService.GetJobs(r.Context(), filters, pagination)
	if err != nil {
		logger.Error("Failed to get company jobs", "error", err, "company", companyName)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve company jobs")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.NewSuccessResponse(jobList))
}

// writeJSONResponse writes a JSON response
func (h *CompanyHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	jobHandler := &JobHandler{} // Reuse the JSON response functionality
	jobHandler.writeJSONResponse(w, statusCode, data)
}

// writeErrorResponse writes an error response
func (h *CompanyHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, models.NewErrorResponse[interface{}](message))
}
