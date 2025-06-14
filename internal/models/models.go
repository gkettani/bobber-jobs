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
	Hash        string    `json:"-"` // for change detection
	FirstSeenAt time.Time `json:"first_seen_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	ExpiredAt   time.Time `json:"expired_at"`
}

type JobReference struct {
	URL        string
	ExternalID string
}
