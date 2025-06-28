package models

import "time"

// JobReference represents a reference to a job posting found during discovery.
// It contains the minimal information needed to locate and identify a job posting.
type JobReference struct {
	// URL is the direct link to the job posting page
	URL string `json:"url"`

	// ExternalID is the unique identifier from the company's system
	ExternalID string `json:"external_id"`

	// CompanyName is the name of the company offering the job
	CompanyName string `json:"company_name"`
}

// IsValid checks if the job reference has all required fields
func (jr *JobReference) IsValid() bool {
	return jr.URL != "" && jr.ExternalID != "" && jr.CompanyName != ""
}

// JobDetails represents complete information about a job posting.
// This is the enriched version created after scraping the job reference.
type JobDetails struct {
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

// IsValid checks if the job details have all required fields
func (jd *JobDetails) IsValid() bool {
	return jd.ExternalID != "" && jd.CompanyName != "" && jd.URL != "" && jd.Title != ""
}
