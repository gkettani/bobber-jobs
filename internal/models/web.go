package models

import "time"

// JobFilters represents query filters for job listing
type JobFilters struct {
	CompanyName string `json:"companyName,omitempty" form:"company"`
	Location    string `json:"location,omitempty" form:"location"`
	Title       string `json:"title,omitempty" form:"title"`
	DateFrom    string `json:"dateFrom,omitempty" form:"dateFrom"`
	DateTo      string `json:"dateTo,omitempty" form:"dateTo"`
	Search      string `json:"search,omitempty" form:"q"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
	Offset   int `json:"-"`
}

// NewPagination creates pagination with defaults
func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// JobList represents a paginated list of jobs
type JobList struct {
	Jobs       []*LightJobDetails `json:"jobs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

// CompanyStats represents statistics for a company
type CompanyStats struct {
	CompanyName string    `json:"companyName" db:"company_name"`
	JobCount    int       `json:"jobCount" db:"job_count"`
	LastUpdated time.Time `json:"lastUpdated" db:"last_updated"`
	ActiveJobs  int       `json:"activeJobs" db:"active_jobs"`
	ExpiredJobs int       `json:"expiredJobs" db:"expired_jobs"`
}

// WebDashboardStatus represents dashboard-specific status information
type WebDashboardStatus struct {
	IsRunning        bool           `json:"isRunning"`
	TotalJobsStored  int64          `json:"totalJobsStored"`
	LastDiscoveryRun time.Time      `json:"lastDiscoveryRun"`
	CompanyStats     []CompanyStats `json:"companyStats"`
	ProcessingRate   float64        `json:"processingRate"`
}

// APIResponse represents a generic API response wrapper
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse[T any](message string) *APIResponse[T] {
	return &APIResponse[T]{
		Success: false,
		Error:   message,
	}
}

type LightJobDetails struct {
	Id          int64     `db:"id" json:"id"`
	CompanyName string    `db:"company_name" json:"companyName"`
	Title       string    `db:"title" json:"title"`
	Location    string    `db:"location" json:"location"`
	FirstSeenAt time.Time `db:"first_seen_at" json:"firstSeenAt"`
	Rank        float64   `db:"rank" json:"rank"`
}
