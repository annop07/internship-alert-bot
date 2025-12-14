package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/annop07/internship-alert-bot/pkg/scraper"
	"github.com/annop07/internship-alert-bot/pkg/storage"
)

func main() {
	printBanner()

	// Initialize storage
	store := storage.NewStorage("data/jobs.json")
	
	// Load existing jobs from storage
	log.Println("\n💾 Step 1: Loading storage...")
	if err := store.Load(); err != nil {
		log.Fatalf("❌ Failed to load storage: %v", err)
	}
	log.Printf("   Currently tracking: %d jobs\n", store.GetJobCount())

	// Create scraper
	s := scraper.NewScraper()

	// Test connection
	log.Println("\n📡 Step 2: Testing connection to JobsDB...")
	if err := s.TestConnection(); err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}

	// Scrape jobs
	log.Println("\n🤖 Step 3: Scraping job listings...")
	scrapedJobs, err := s.ScrapeJobs()
	if err != nil {
		log.Fatalf("❌ Scraping failed: %v", err)
	}

	// Compare with storage to find new jobs
	log.Println("\n🔍 Step 4: Comparing with storage...")
	newJobs := store.GetNewJobs(scrapedJobs)
	
	log.Println("\n" + strings.Repeat("=", 70))
	log.Printf("📊 RESULTS\n")
	log.Println(strings.Repeat("=", 70))
	log.Printf("   Total jobs found:  %d\n", len(scrapedJobs))
	log.Printf("   Already tracked:   %d\n", store.GetJobCount())
	log.Printf("   🔥 NEW jobs:       %d\n", len(newJobs))
	log.Println(strings.Repeat("=", 70))

	// Display results
	if len(newJobs) == 0 {
		log.Println("\n✅ No new jobs found. All jobs are already tracked.")
		log.Println("💡 This is good! It means the system is working correctly.")
		log.Println("   Run again later to check for new postings.")
	} else {
		log.Printf("\n🎉 Found %d NEW job(s)!\n", len(newJobs))
		log.Println(strings.Repeat("=", 70))
		
		// Print only new jobs
		for i, job := range newJobs {
			printJob(i+1, job)
		}
		
		// Add new jobs to storage
		log.Println("\n💾 Step 5: Updating storage...")
		store.AddJobs(newJobs)
		
		if err := store.Save(); err != nil {
			log.Fatalf("❌ Failed to save storage: %v", err)
		}
		
		log.Printf("✅ Successfully added %d new job(s) to storage\n", len(newJobs))
		log.Printf("📊 Now tracking: %d total jobs\n", store.GetJobCount())
	}

	// Print summary
	printStorageSummary(store)
	
	log.Println("\n✅ Phase 2 Complete!")
	log.Println("🎯 Next time you run, only NEW jobs will be displayed!")
	log.Println("💡 Try running again in a few minutes to test the storage system.")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║         🤖 INTERNSHIP ALERT BOT - Phase 2               ║
║            Now with Job Memory! 💾                       ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printJob(index int, job interface{}) {
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
	
	fmt.Printf("\n🆕 NEW Job #%d\n", index)
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

func printStorageSummary(store *storage.Storage) {
	stats := store.GetStats()
	
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📈 STORAGE SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	
	fmt.Printf("Total Jobs Tracked:    %d\n", stats["total_jobs"])
	fmt.Printf("Unique Companies:      %d\n", stats["unique_companies"])
	fmt.Printf("Unique Locations:      %d\n", stats["unique_locations"])
	
	// Show top companies
	if companies, ok := stats["companies"].(map[string]int); ok && len(companies) > 0 {
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
	
	// Show top locations
	if locations, ok := stats["locations"].(map[string]int); ok && len(locations) > 0 {
		fmt.Println("\n📍 Top Locations:")
		count := 0
		for location, num := range locations {
			if count >= 5 {
				break
			}
			fmt.Printf("   • %s (%d position%s)\n", location, num, pluralize(num))
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