package models

import "time"

type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	PostedDate  string    `json:"posted_date"`
	ScrapedAt   time.Time `json:"scraped_at"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category"`
}

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
