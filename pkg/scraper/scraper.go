package scraper

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/annop07/internship-alert-bot/pkg/models"
)

const (
	maxRetries     = 3
	retryDelay     = 2 * time.Second
	requestDelay   = 1 * time.Second
	timeoutSeconds = 30
)

// Scraper handles job scraping from JobsDB
type Scraper struct {
	client   *http.Client
	baseURL  string
	category string
}

// NewScraper creates a new Scraper instance for a specific job category
func NewScraper(category string) *Scraper {
	var searchKeyword string
	switch category {
	case "backend":
		searchKeyword = "backend%20internship"
	case "frontend":
		searchKeyword = "frontend%20internship"
	case "fullstack":
		searchKeyword = "fullstack%20internship"
	default:
		searchKeyword = "internship" // fallback
	}

	baseURL := fmt.Sprintf("https://th.jobsdb.com/th/search-jobs/%s/1", searchKeyword)

	return &Scraper{
		client: &http.Client{
			Timeout: timeoutSeconds * time.Second,
		},
		baseURL:  baseURL,
		category: category,
	}
}

// ScrapeJobs fetches and parses job listings with retry logic
func (s *Scraper) ScrapeJobs() ([]*models.Job, error) {
	log.Println("🔍 Starting to scrape jobs from JobsDB...")

	doc, err := s.fetchHTMLWithRetry(s.baseURL, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs after %d retries: %w", maxRetries, err)
	}

	if err := s.validateHTML(doc); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	jobs := s.extractJobs(doc)
	log.Printf("✅ Successfully extracted %d jobs\n", len(jobs))

	return jobs, nil
}

// fetchHTMLWithRetry fetches HTML with exponential backoff retry
func (s *Scraper) fetchHTMLWithRetry(url string, maxRetries int) (*goquery.Document, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		s.setHeaders(req)

		if attempt > 1 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * retryDelay
			log.Printf("⏳ Waiting %v before retry %d/%d...", delay, attempt, maxRetries)
			time.Sleep(delay)
		} else if attempt == 1 {
			time.Sleep(requestDelay)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt, maxRetries, err)
			log.Printf("⚠️  %v", lastErr)
			continue
		}

		// Check status code
		if resp.StatusCode != 200 {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code %d (attempt %d/%d)", resp.StatusCode, attempt, maxRetries)
			log.Printf("⚠️  %v", lastErr)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to parse HTML (attempt %d/%d): %w", attempt, maxRetries, err)
			log.Printf("⚠️  %v", lastErr)
			continue
		}

		log.Println("✅ Page fetched successfully, parsing HTML...")
		return doc, nil
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// setHeaders sets browser-like headers
func (s *Scraper) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "th-TH,th;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")
}

func (s *Scraper) validateHTML(doc *goquery.Document) error {
	jobCards := doc.Find("article[data-automation='normalJob']")
	if jobCards.Length() == 0 {
		return fmt.Errorf("HTML structure changed: no job cards found with selector 'article[data-automation=normalJob]'")
	}

	firstCard := jobCards.First()
	if firstCard.Find("a[data-automation='jobTitle']").Length() == 0 {
		return fmt.Errorf("HTML structure changed: jobTitle selector not found")
	}
	if firstCard.Find("a[data-automation='jobCompany']").Length() == 0 {
		return fmt.Errorf("HTML structure changed: jobCompany selector not found")
	}

	return nil
}

func (s *Scraper) extractJobs(doc *goquery.Document) []*models.Job {
	var jobs []*models.Job

	doc.Find("article[data-automation='normalJob']").Each(func(i int, selection *goquery.Selection) {
		jobID, _ := selection.Attr("data-job-id")
		title := strings.TrimSpace(selection.Find("a[data-automation='jobTitle']").Text())
		company := strings.TrimSpace(selection.Find("a[data-automation='jobCompany']").Text())
		location := strings.TrimSpace(selection.Find("a[data-automation='jobLocation']").Text())

		urlPath, exists := selection.Find("a[data-automation='jobTitle']").Attr("href")
		url := ""
		if exists && urlPath != "" {
			if strings.Contains(urlPath, "?") {
				urlPath = strings.Split(urlPath, "?")[0]
			}

			if strings.HasPrefix(urlPath, "http") {
				url = urlPath
			} else {
				url = "https://th.jobsdb.com" + urlPath
			}
		}

		postedDate := strings.TrimSpace(selection.Find("span[data-automation='jobListingDate']").Text())
		if len(postedDate) > 0 && len(postedDate)%2 == 0 {
			half := len(postedDate) / 2
			firstHalf := postedDate[:half]
			secondHalf := postedDate[half:]
			if firstHalf == secondHalf {
				postedDate = firstHalf
			}
		}

		description := strings.TrimSpace(selection.Find("span[data-automation='jobShortDescription']").Text())

		if title != "" && company != "" && url != "" {
			job := models.NewJob(title, company, location, url, postedDate)
			job.ID = jobID
			job.Description = description
			job.Category = s.category

			jobs = append(jobs, job)

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

func (s *Scraper) TestConnection() error {
	log.Println("🔗 Testing connection to JobsDB...")

	_, err := s.fetchHTMLWithRetry(s.baseURL, 2)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	log.Println("✅ Connection test successful!")
	return nil
}
