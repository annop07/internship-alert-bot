package models

import "time"

// Job represents a job posting
type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	PostedDate  string    `json:"posted_date"`
	ScrapedAt   time.Time `json:"scraped_at"`
	Description string    `json:"description,omitempty"`
}

// NewJob creates a new Job instance
func NewJob(title, company, location, url, postedDate string) *Job {
	return &Job{
		Title:      title,
		Company:    company,
		Location:   location,
		URL:        url,
		PostedDate: postedDate,
		ScrapedAt:  time.Now(),
	}
}