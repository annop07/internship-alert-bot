package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/annop07/internship-alert-bot/pkg/models"
)

// Scraper handles job scraping from JobsDB
type Scraper struct {
	client  *http.Client
	baseURL string
}

// NewScraper creates a new Scraper instance
func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		// URL สำหรับค้นหา internship
		baseURL: "https://th.jobsdb.com/th/search-jobs/backend%20internship/1",
	}
}

// ScrapeJobs fetches and parses job listings
func (s *Scraper) ScrapeJobs() ([]*models.Job, error) {
	log.Println("🔍 Starting to scrape jobs from JobsDB...")

	// Create request
	req, err := http.NewRequest("GET", s.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "th-TH,th;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Println("✅ Page fetched successfully, parsing HTML...")

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract jobs
	jobs := s.extractJobs(doc)
	log.Printf("✅ Successfully extracted %d jobs\n", len(jobs))

	return jobs, nil
}

// extractJobs parses job listings from HTML document
func (s *Scraper) extractJobs(doc *goquery.Document) []*models.Job {
	var jobs []*models.Job

	// Find all job cards using the correct selector
	doc.Find("article[data-automation='normalJob']").Each(func(i int, selection *goquery.Selection) {

		// Extract Job ID
		jobID, _ := selection.Attr("data-job-id")

		// Extract Title
		title := strings.TrimSpace(selection.Find("a[data-automation='jobTitle']").Text())

		// Extract Company
		company := strings.TrimSpace(selection.Find("a[data-automation='jobCompany']").Text())

		// Extract Location
		location := strings.TrimSpace(selection.Find("a[data-automation='jobLocation']").Text())

		// Extract URL
		urlPath, exists := selection.Find("a[data-automation='jobTitle']").Attr("href")
		url := ""
		if exists && urlPath != "" {
			// Clean URL - remove query parameters for cleaner link
			// Example: /th/job/88662455?type=standard... → /th/job/88662455
			if strings.Contains(urlPath, "?") {
				urlPath = strings.Split(urlPath, "?")[0]
			}

			if strings.HasPrefix(urlPath, "http") {
				url = urlPath
			} else {
				url = "https://th.jobsdb.com" + urlPath
			}
		}

		// Extract Posted Date
		postedDate := strings.TrimSpace(selection.Find("span[data-automation='jobListingDate']").Text())

		// Extract Description (optional)
		description := strings.TrimSpace(selection.Find("span[data-automation='jobShortDescription']").Text())

		// Validate minimum required data
		if title != "" && company != "" && url != "" {
			job := models.NewJob(title, company, location, url, postedDate)
			job.ID = jobID
			job.Description = description

			jobs = append(jobs, job)

			// Debug log for first few jobs
			if i < 3 {
				log.Printf("   [%d] %s at %s", i+1, title, company)
			}
		} else {
			log.Printf("⚠️  Skipped job card %d - missing required data (title: %t, company: %t, url: %t)",
				i, title != "", company != "", url != "")
		}
	})

	return jobs
}

// TestConnection checks if the scraper can reach JobsDB
func (s *Scraper) TestConnection() error {
	log.Println("🔗 Testing connection to JobsDB...")

	// Create request
	req, err := http.NewRequest("GET", s.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic browser (same as ScrapeJobs)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "th-TH,th;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Println("✅ Connection test successful!")
	return nil
}
