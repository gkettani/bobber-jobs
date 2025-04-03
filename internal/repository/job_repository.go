package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gkettani/bobber-the-swe/internal/db"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/jmoiron/sqlx"
)

type JobRepository interface {
	Insert(ctx context.Context, job *models.Job) error
	Upsert(ctx context.Context, job *models.Job) error
	BulkInsert(ctx context.Context, jobs []*models.Job) error
	FindByID(ctx context.Context, id int64) (*models.Job, error)
}

type jobRepository struct {
	db        *sqlx.DB
	batchSize int
}

func NewJobRepository(client *db.DBClient, batchSize int) *jobRepository {
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	return &jobRepository{
		db:        client.GetConnection(),
		batchSize: batchSize,
	}
}

func (r *jobRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *jobRepository) Insert(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			title, description, company_name, location, url, external_id, hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id`

	err := r.db.QueryRowxContext(
		ctx,
		query,
		job.Title,
		job.Description,
		job.CompanyName,
		job.Location,
		job.URL,
		job.ExternalID,
		job.Hash,
	).Scan(&job.ID)

	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	return nil
}

func (r *jobRepository) Upsert(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			title, description, company_name, location, url, external_id, hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) ON CONFLICT (external_id) DO UPDATE 
		SET
			last_seen_at = NOW(),
			description = EXCLUDED.description,
			hash = EXCLUDED.hash
		WHERE jobs.hash <> EXCLUDED.hash
		RETURNING id`

	err := r.db.QueryRowxContext(
		ctx,
		query,
		job.Title,
		job.Description,
		job.CompanyName,
		job.Location,
		job.URL,
		job.ExternalID,
		job.Hash,
	).Scan(&job.ID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to insert or update job: %w", err)
	}

	return nil
}

func (r *jobRepository) BulkInsert(ctx context.Context, jobs []*models.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	return r.WithTransaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		// Process in batches to avoid hitting statement size limits
		for i := 0; i < len(jobs); i += r.batchSize {
			end := min(i+r.batchSize, len(jobs))

			batch := jobs[i:end]

			placeholders := make([]string, len(batch))
			values := make([]any, 0, len(batch)*8)

			for j, job := range batch {
				// Calculate placeholder position
				pos := j * 7
				placeholders[j] = fmt.Sprintf(
					"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
					pos+1, pos+2, pos+3, pos+4, pos+5, pos+6, pos+7,
				)

				values = append(values,
					job.Title,
					job.Description,
					job.CompanyName,
					job.Location,
					job.URL,
					job.ExternalID,
					job.Hash,
				)
			}

			query := fmt.Sprintf(`
				INSERT INTO jobs (
					title, description, company_name, location, url, external_id, hash
				) VALUES %s
				ON CONFLICT (external_id) DO UPDATE 
				SET
					last_seen_at = NOW(),
					description = EXCLUDED.description,
					hash = EXCLUDED.hash
				WHERE jobs.hash <> EXCLUDED.hash
				RETURNING id`, strings.Join(placeholders, ","))

			rows, err := tx.QueryxContext(ctx, query, values...)
			if err != nil {
				return fmt.Errorf("failed to bulk insert jobs: %w", err)
			}

			// Assign IDs to the original job objects
			j := i
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err != nil {
					rows.Close()
					return fmt.Errorf("failed to scan returned job ID: %w", err)
				}
				// Only set ID if it was successfully inserted (might be less than batch size due to ON CONFLICT)
				if j < len(jobs) {
					jobs[j].ID = id
					j++
				}
			}

			if err := rows.Err(); err != nil {
				return fmt.Errorf("error iterating through result rows: %w", err)
			}

			rows.Close()
		}

		return nil
	})
}

func (r *jobRepository) FindByID(ctx context.Context, id int64) (*models.Job, error) {
	query := `SELECT id, title, description, company_name, external_id, location, url FROM jobs WHERE id = $1`

	var job models.Job
	err := r.db.GetContext(ctx, &job, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("job not found")
		}
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	return &job, nil
}
