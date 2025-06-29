package query

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/gkettani/bobber-the-swe/internal/services"
	"github.com/jmoiron/sqlx"
)

type jobQueryService struct {
	db *sqlx.DB
}

// NewJobQueryService creates a new job query service
func NewJobQueryService() services.JobQueryService {
	return &jobQueryService{
		db: db.GetDBClient().GetConnection(),
	}
}

// GetJobs retrieves jobs with filtering and pagination
func (s *jobQueryService) GetJobs(ctx context.Context, filters *models.JobFilters, pagination *models.Pagination) (*models.JobList, error) {
	// Build the WHERE clause
	whereClause, args := s.buildWhereClause(filters)

	// Count total jobs matching the filters
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM jobs %s", whereClause)
	var total int64
	err := s.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Build the main query with pagination
	query := fmt.Sprintf(`
		SELECT id, company_name, title, location, first_seen_at
		FROM jobs %s
		ORDER BY last_seen_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pagination.PageSize, pagination.Offset)

	var jobs []*models.LightJobDetails
	err = s.db.SelectContext(ctx, &jobs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))

	return &models.JobList{
		Jobs:       jobs,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetJobByID retrieves a specific job by ID
func (s *jobQueryService) GetJobByID(ctx context.Context, id int64) (*models.JobDetails, error) {
	query := `
		SELECT id, external_id, company_name, url, title, location, description,
		       first_seen_at, last_seen_at
		FROM jobs 
		WHERE id = $1
	`

	var job models.JobDetails
	err := s.db.GetContext(ctx, &job, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// SearchJobs performs full-text search on jobs
func (s *jobQueryService) SearchJobs(ctx context.Context, searchQuery string, pagination *models.Pagination) (*models.JobList, error) {
	if searchQuery == "" {
		return s.GetJobs(ctx, &models.JobFilters{}, pagination)
	}

	// Count total jobs matching the search
	countQuery := `
		SELECT COUNT(*) FROM jobs 
		WHERE search_vector @@ plainto_tsquery('english', $1)
	`
	var total int64
	err := s.db.GetContext(ctx, &total, countQuery, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get the jobs with search ranking
	query := `
		SELECT id, company_name, title, location, first_seen_at,
		       ts_rank(search_vector, plainto_tsquery('english', $1)) as rank
		FROM jobs 
		WHERE search_vector @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC, last_seen_at DESC
		LIMIT $2 OFFSET $3
	`

	var jobs []*models.LightJobDetails
	err = s.db.SelectContext(ctx, &jobs, query, searchQuery, pagination.PageSize, pagination.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search jobs: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))

	return &models.JobList{
		Jobs:       jobs,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetCompanyStats retrieves statistics for all companies
func (s *jobQueryService) GetCompanyStats(ctx context.Context) ([]models.CompanyStats, error) {
	query := `
		SELECT 
			company_name,
			COUNT(*) as job_count,
			MAX(last_seen_at) as last_updated,
			COUNT(CASE WHEN expired_at IS NULL THEN 1 END) as active_jobs,
			COUNT(CASE WHEN expired_at IS NOT NULL THEN 1 END) as expired_jobs
		FROM jobs 
		GROUP BY company_name
		ORDER BY job_count DESC
	`

	var stats []models.CompanyStats
	err := s.db.SelectContext(ctx, &stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get company stats: %w", err)
	}

	return stats, nil
}

// GetJobCountByCompany returns job counts grouped by company
func (s *jobQueryService) GetJobCountByCompany(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT company_name, COUNT(*) as count
		FROM jobs 
		GROUP BY company_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get job counts by company: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var company string
		var count int
		if err := rows.Scan(&company, &count); err != nil {
			return nil, fmt.Errorf("failed to scan job count: %w", err)
		}
		counts[company] = count
	}

	return counts, nil
}

// GetTotalJobCount returns the total number of jobs in the database
func (s *jobQueryService) GetTotalJobCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM jobs")
	if err != nil {
		return 0, fmt.Errorf("failed to get total job count: %w", err)
	}
	return count, nil
}

// GetRecentJobs returns the most recently discovered jobs
func (s *jobQueryService) GetRecentJobs(ctx context.Context, limit int) ([]*models.LightJobDetails, error) {
	query := `
		SELECT id, company_name, title, location,
		       first_seen_at
		FROM jobs 
		ORDER BY first_seen_at DESC
		LIMIT $1
	`

	var jobs []*models.LightJobDetails
	err := s.db.SelectContext(ctx, &jobs, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent jobs: %w", err)
	}

	return jobs, nil
}

// buildWhereClause builds the WHERE clause and arguments for filtering
func (s *jobQueryService) buildWhereClause(filters *models.JobFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.CompanyName != "" {
		conditions = append(conditions, fmt.Sprintf("company_name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.CompanyName+"%")
		argIndex++
	}

	if filters.Location != "" {
		conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Location+"%")
		argIndex++
	}

	if filters.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Title+"%")
		argIndex++
	}

	if filters.DateFrom != "" {
		if dateFrom, err := time.Parse("2006-01-02", filters.DateFrom); err == nil {
			conditions = append(conditions, fmt.Sprintf("first_seen_at >= $%d", argIndex))
			args = append(args, dateFrom)
			argIndex++
		}
	}

	if filters.DateTo != "" {
		if dateTo, err := time.Parse("2006-01-02", filters.DateTo); err == nil {
			conditions = append(conditions, fmt.Sprintf("first_seen_at <= $%d", argIndex))
			args = append(args, dateTo.Add(24*time.Hour)) // Include the entire day
			argIndex++
		}
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("search_vector @@ plainto_tsquery('english', $%d)", argIndex))
		args = append(args, filters.Search)
		argIndex++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
