package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/annop07/internship-alert-bot/pkg/models"
)

type Storage struct {
	filePath string
	jobs     map[string]*models.Job // Key:Job ID
}

func NewStorage(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
		jobs:     make(map[string]*models.Job),
	}
}

// Load reads jobs from storage file
func (s *Storage) Load() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		log.Println("📂 No existing storage file, starting fresh")
		return nil // Not an error - first run
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	var jobs []*models.Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return fmt.Errorf("failed to parse storage file: %w", err)
	}


	s.jobs = make(map[string]*models.Job)
	for _, job := range jobs {
		s.jobs[job.ID] = job
	}

	log.Printf("📂 Loaded %d jobs from storage\n", len(s.jobs))
	return nil
}

func (s *Storage) Save() error {
	// Convert map to slice
	jobs := make([]*models.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal jobs: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	log.Printf("💾 Saved %d jobs to storage\n", len(s.jobs))
	return nil
}

func (s *Storage) IsNewJob(job *models.Job) bool {
	_, exists := s.jobs[job.ID]
	return !exists
}

func (s *Storage) AddJob(job *models.Job) {
	s.jobs[job.ID] = job
}

func (s *Storage) AddJobs(jobs []*models.Job) {
	for _, job := range jobs {
		s.jobs[job.ID] = job
	}
}

func (s *Storage) GetNewJobs(scrapedJobs []*models.Job) []*models.Job {
	var newJobs []*models.Job

	for _, job := range scrapedJobs {
		if s.IsNewJob(job) {
			newJobs = append(newJobs, job)
		}
	}

	return newJobs
}


func (s *Storage) GetAllJobs() []*models.Job {
	jobs := make([]*models.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}


func (s *Storage) GetJobCount() int {
	return len(s.jobs)
}


func (s *Storage) CleanOldJobs(days int) int {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	removed := 0

	for id, job := range s.jobs {
		if job.ScrapedAt.Before(cutoffTime) {
			delete(s.jobs, id)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("🧹 Cleaned %d old jobs (older than %d days)\n", removed, days)
	}

	return removed
}


func (s *Storage) GetStats() map[string]interface{} {
	companies := make(map[string]int)
	locations := make(map[string]int)

	for _, job := range s.jobs {
		companies[job.Company]++
		if job.Location != "" {
			locations[job.Location]++
		}
	}

	return map[string]interface{}{
		"total_jobs":       len(s.jobs),
		"unique_companies": len(companies),
		"unique_locations": len(locations),
		"companies":        companies,
		"locations":        locations,
	}
}


func (s *Storage) GetRecentJobs(limit int) []*models.Job {
	allJobs := s.GetAllJobs()

	if len(allJobs) == 0 {
		return []*models.Job{}
	}


	count := limit
	if count > len(allJobs) {
		count = len(allJobs)
	}

	return allJobs[:count]
}
