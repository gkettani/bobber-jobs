package models

import "time"

type Link struct {
	URL string
}

type Job struct {
	ID          int64     `db:"id"`
	ExternalID  string    `db:"external_id"`
	CompanyName string    `db:"company_name"`
	URL         string    `db:"job_url"`
	Title       string    `db:"title"`
	Location    string    `db:"location"`
	Description string    `db:"description"`
	Hash        string    `json:"-"` // Used for tracking changes
	PostedAt    time.Time `json:"posted_at"`
	FirstSeenAt time.Time `json:"first_seen_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	ExpiredAt   time.Time `json:"expired_at"`
}

type JobListing struct {
	URL        string
	ExternalID string
}
