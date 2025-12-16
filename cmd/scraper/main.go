package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/annop07/internship-alert-bot/pkg/logger"
	"github.com/annop07/internship-alert-bot/pkg/notifier"
	"github.com/annop07/internship-alert-bot/pkg/scraper"
	"github.com/annop07/internship-alert-bot/pkg/storage"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

var scheduled bool

// Job categories to scrape
var categories = []string{"backend", "frontend", "fullstack"}

// Category emojis for LINE
var categoryEmojis = map[string]string{
	"backend":   "🔵",
	"frontend":  "🟢",
	"fullstack": "🟡",
}

func main() {
	// Parse command-line flags
	flag.BoolVar(&scheduled, "scheduled", false, "Run in scheduled mode (every 15 minutes)")
	flag.Parse()

	printBanner()

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using system environment variables")
	}

	if scheduled {
		runScheduler()
	} else {
		runScraperWithRecovery()
	}
}

func runScheduler() {
	logger.Info("")
	logger.Info("⏰ Starting scheduler mode...")
	logger.Info("📅 Bot will run every 15 minutes")
	logger.Info("🛑 Press Ctrl+C to stop")
	logger.Info("")

	// Create cron scheduler
	c := cron.New()

	// Schedule: every 15 minutes
	_, err := c.AddFunc("*/15 * * * *", func() {
		logger.LogScheduledRun()
		runScraperWithRecovery()
	})

	if err != nil {
		logger.Fatal("Failed to setup scheduler: %v", err)
	}

	// Start the scheduler
	c.Start()

	// Run immediately on startup
	logger.Info("🚀 Running initial scan...")
	runScraperWithRecovery()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan

	logger.Info("")
	logger.Info("")
	logger.Info("🛑 Shutdown signal received...")
	logger.Info("⏳ Stopping scheduler...")

	// Stop the scheduler
	ctx := c.Stop()
	<-ctx.Done()

	logger.Success("Scheduler stopped gracefully")
	logger.Info("👋 Goodbye!")
}

// runScraperWithRecovery wraps runScraper with panic recovery
func runScraperWithRecovery() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("💥 PANIC RECOVERED: %v", r)

			// Try to send error notification
			if err := sendErrorNotification(fmt.Sprintf("Bot panicked: %v", r)); err != nil {
				logger.Error("Failed to send panic notification: %v", err)
			}

			// Continue running if in scheduler mode
			if scheduled {
				logger.Info("Scheduler will continue with next run...")
			}
		}
	}()

	runScraper()
}

// sendErrorNotification sends critical error alerts
func sendErrorNotification(errorMsg string) error {
	discordWebhook := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhook == "" {
		return fmt.Errorf("no Discord webhook configured")
	}

	// Log the error locally
	logger.Error("Critical error notification: %s", errorMsg)

	// TODO: Implement proper error notification via Discord
	// For now, just log it
	return nil
}

func runScraper() {
	// Initialize LINE Bot notifier (Optional)
	lineChannelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	lineChannelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	lineUserID := os.Getenv("LINE_USER_ID")

	var lineBot *notifier.LineBotNotifier
	if lineChannelSecret != "" && lineChannelToken != "" && lineUserID != "" {
		var err error
		lineBot, err = notifier.NewLineBotNotifier(lineChannelSecret, lineChannelToken, lineUserID)
		if err != nil {
			log.Printf("⚠️ Failed to initialize LINE Bot: %v", err)
		}
	} else {
		log.Println("⚠️ LINE credentials missing. Skipping LINE notifications.")
	}

	// Test LINE connection if available (no message sent to user in production)
	if lineBot != nil {
		log.Println("✅ LINE Bot initialized successfully")
	}

	// Scrape each category
	for _, category := range categories {
		emoji := categoryEmojis[category]
		log.Printf("\n%s Scraping %s internships...\n", emoji, strings.Title(category))

		// Create scraper for this category
		s := scraper.NewScraper(category)

		// Initialize storage for this category
		storageFile := fmt.Sprintf("data/%s_jobs.json", category)
		store := storage.NewStorage(storageFile)

		// Load existing jobs
		if err := store.Load(); err != nil {
			log.Printf("⚠️  No existing storage for %s, creating new", category)
		}
		log.Printf("   Currently tracking: %d %s jobs", store.GetJobCount(), category)

		// Scrape jobs
		scrapedJobs, err := s.ScrapeJobs()
		if err != nil {
			log.Printf(" ❌ Failed to scrape %s jobs: %v", category, err)
			continue
		}

		// Find new jobs
		newJobs := store.GetNewJobs(scrapedJobs)

		log.Printf("   Found %d total, %d new %s jobs", len(scrapedJobs), len(newJobs), category)

		// Send LINE notification if there are new jobs
		if lineBot != nil && len(newJobs) > 0 {
			log.Printf("\n📱 Sending %s notifications to LINE...", category)
			if err := lineBot.SendMultipleJobsAlert(newJobs); err != nil {
				log.Printf("⚠️  Failed to send LINE notifications: %v", err)
			} else {
				log.Printf("✅ Sent %d %s job(s) to LINE!", len(newJobs), category)
			}
		}

		// Save new jobs
		if len(newJobs) > 0 {
			store.AddJobs(newJobs)
			if err := store.Save(); err != nil {
				log.Printf("❌ Failed to save %s storage: %v", category, err)
			} else {
				log.Printf("💾 Saved %d new %s jobs", len(newJobs), category)
			}
		}
	}

	log.Println("\n" + strings.Repeat("=", 70))
	log.Println("✅ Multi-Category Scraping Complete!")
	log.Println("🎯 Checked: Backend, Frontend, Fullstack")
	log.Println("💡 Check your LINE for category-specific alerts.")
	log.Println(strings.Repeat("=", 70))
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║         🤖 INTERNSHIP ALERT BOT - Multi-Category       ║
║            Backend | Frontend | Fullstack  🔵🟢🟡        ║
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
		Category    string `json:"category"`
	}

	jobData, _ := json.Marshal(job)
	var jobInfo JobInfo
	json.Unmarshal(jobData, &jobInfo)

	emoji := categoryEmojis[jobInfo.Category]
	fmt.Printf("\n%s NEW %s Job #%d\n", emoji, strings.Title(jobInfo.Category), index)
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
