package models

import "time"

// JobReference represents a reference to a job posting found during discovery.
// It contains the minimal information needed to locate and identify a job posting.
type JobReference struct {
	// URL is the direct link to the job posting page
	URL string

	// ExternalID is the unique identifier from the company's system
	ExternalID string

	// CompanyName is the name of the company offering the job
	CompanyName string
}

// IsValid checks if the job reference has all required fields
func (jr *JobReference) IsValid() bool {
	return jr.URL != "" && jr.ExternalID != "" && jr.CompanyName != ""
}

// JobDetails represents complete information about a job posting.
// This is the enriched version created after scraping the job reference.
type JobDetails struct {
	ID          int64     `db:"id" json:"id"`
	ExternalID  string    `db:"external_id" json:"externalId"`
	CompanyName string    `db:"company_name" json:"companyName"`
	URL         string    `db:"url" json:"url"`
	Title       string    `db:"title" json:"title"`
	Location    string    `db:"location" json:"location"`
	Description string    `db:"description" json:"description"`
	Hash        string    `json:"-"` // for change detection
	FirstSeenAt time.Time `db:"first_seen_at" json:"firstSeenAt"`
	LastSeenAt  time.Time `db:"last_seen_at" json:"lastSeenAt"`
	ExpiredAt   time.Time `db:"expired_at" json:"expiredAt"`
}

// IsValid checks if the job details have all required fields
func (jd *JobDetails) IsValid() bool {
	return jd.ExternalID != "" && jd.CompanyName != "" && jd.URL != "" && jd.Title != ""
}
