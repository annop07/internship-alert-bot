package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/annop07/internship-alert-bot/pkg/scraper"
)

func main() {
	printBanner()

	// Create scraper
	s := scraper.NewScraper()

	// Test connection first
	log.Println("\n📡 Step 1: Testing connection to JobsDB...")
	if err := s.TestConnection(); err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}

	// Scrape jobs
	log.Println("\n🤖 Step 2: Scraping job listings...")
	jobs, err := s.ScrapeJobs()
	if err != nil {
		log.Fatalf("❌ Scraping failed: %v", err)
	}

	// Print results
	log.Println("\n" + strings.Repeat("=", 70))
	log.Printf("📊 RESULTS: Found %d internship positions\n", len(jobs))
	log.Println(strings.Repeat("=", 70))

	if len(jobs) == 0 {
		log.Println("⚠️  No jobs found.")
		log.Println("💡 This could mean:")
		log.Println("   - No internships are currently posted")
		log.Println("   - JobsDB changed their HTML structure")
		log.Println("   - The website is blocking our scraper")
		return
	}

	// Print jobs in a nice format
	for i, job := range jobs {
		printJob(i+1, job)
	}

	// Save to JSON file for inspection
	if err := saveToJSON(jobs, "jobs_output.json"); err != nil {
		log.Printf("⚠️  Could not save JSON file: %v", err)
	} else {
		log.Println("\n💾 Jobs saved to jobs_output.json")
	}

	// Print summary
	printSummary(jobs)

	log.Println("\n✅ Phase 1 Complete!")
	log.Println("🎯 Next: Move to Phase 2 - Add storage to track new jobs")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║         🤖 INTERNSHIP ALERT BOT - Phase 1               ║
║            JobsDB Scraper Demo                          ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printJob(index int, job interface{}) {
	// Type assertion to access job fields
	type JobInfo struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Company     string `json:"company"`
		Location    string `json:"location"`
		URL         string `json:"url"`
		PostedDate  string `json:"posted_date"`
		Description string `json:"description,omitempty"`
	}

	jobData, _ := json.Marshal(job)
	var jobInfo JobInfo
	json.Unmarshal(jobData, &jobInfo)

	fmt.Printf("\n📌 Job #%d\n", index)
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("   🏷️  ID:       %s\n", jobInfo.ID)
	fmt.Printf("   💼 Title:    %s\n", jobInfo.Title)
	fmt.Printf("   🏢 Company:  %s\n", jobInfo.Company)
	fmt.Printf("   📍 Location: %s\n", jobInfo.Location)
	fmt.Printf("   📅 Posted:   %s\n", jobInfo.PostedDate)

	if jobInfo.Description != "" {
		desc := jobInfo.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		fmt.Printf("   📝 Preview:  %s\n", desc)
	}

	fmt.Printf("   🔗 URL:      %s\n", jobInfo.URL)
}

func printSummary(jobs interface{}) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📈 SUMMARY")
	fmt.Println(strings.Repeat("=", 70))

	// Count jobs by company
	companies := make(map[string]int)

	jobData, _ := json.Marshal(jobs)
	var jobList []struct {
		Company string `json:"company"`
	}
	json.Unmarshal(jobData, &jobList)

	for _, job := range jobList {
		companies[job.Company]++
	}

	fmt.Printf("Total Jobs:        %d\n", len(jobList))
	fmt.Printf("Unique Companies:  %d\n", len(companies))

	if len(companies) > 0 {
		fmt.Println("\n🏆 Top Companies:")
		count := 0
		for company, num := range companies {
			if count >= 5 {
				break
			}
			fmt.Printf("   • %s (%d position%s)\n", company, num, pluralize(num))
			count++
		}
	}
}

func pluralize(count int) string {
	if count > 1 {
		return "s"
	}
	return ""
}

func saveToJSON(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}
