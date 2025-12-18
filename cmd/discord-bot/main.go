package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/annop07/internship-alert-bot/pkg/models"
	"github.com/annop07/internship-alert-bot/pkg/storage"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const jobsPerPage = 10

var s *discordgo.Session

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found")
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ DISCORD_BOT_TOKEN not set")
	}

	// Create Discord session
	var err error
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("❌ Failed to create Discord session: %v", err)
	}

	s.AddHandler(ready)
	s.AddHandler(interactionCreate)

	// Open connection
	if err := s.Open(); err != nil {
		log.Fatalf("❌ Failed to open Discord connection: %v", err)
	}
	defer s.Close()

	registerCommands()

	fmt.Println("🤖 Discord Bot is running!")
	fmt.Println("   Commands: /backend, /frontend, /fullstack, /stats")
	fmt.Println("   Press Ctrl+C to stop")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\n👋 Shutting down...")
	removeCommands()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("✅ Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "backend",
		Description: "ดูตำแหน่ง Backend Internship ล่าสุด",
	},
	{
		Name:        "frontend",
		Description: "ดูตำแหน่ง Frontend Internship ล่าสุด",
	},
	{
		Name:        "fullstack",
		Description: "ดูตำแหน่ง Fullstack Internship ล่าสุด",
	},
	{
		Name:        "stats",
		Description: "ดูสถิติงาน Internship ทั้งหมด",
	},
}

var registeredCommands []*discordgo.ApplicationCommand

func registerCommands() {
	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, cmd := range commands {
		rcmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("⚠️  Cannot create command '%v': %v", cmd.Name, err)
		} else {
			registeredCommands[i] = rcmd
			log.Printf("✅ Registered command: /%s", cmd.Name)
		}
	}
}

func removeCommands() {
	for _, cmd := range registeredCommands {
		if cmd != nil {
			s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
		}
	}
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		handleButton(s, i)
	}
}

func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	log.Printf("📥 Command received: /%s", data.Name)

	switch data.Name {
	case "backend":
		sendJobPage(s, i, "backend", 1, false)
	case "frontend":
		sendJobPage(s, i, "frontend", 1, false)
	case "fullstack":
		sendJobPage(s, i, "fullstack", 1, false)
	case "stats":
		handleStatsCommand(s, i)
	}
}

func handleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse button ID: "jobs:backend:2"
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 3 || parts[0] != "jobs" {
		return
	}

	category := parts[1]
	page, _ := strconv.Atoi(parts[2])

	log.Printf("📥 Button clicked: %s page %d", category, page)
	sendJobPage(s, i, category, page, true)
}

func sendJobPage(s *discordgo.Session, i *discordgo.InteractionCreate, category string, page int, isUpdate bool) {
	emoji := getCategoryEmoji(category)

	// Load jobs from storage
	storageFile := fmt.Sprintf("data/%s_jobs.json", category)
	store := storage.NewStorage(storageFile)

	if err := store.Load(); err != nil {
		respondWithError(s, i, "ไม่สามารถดึงข้อมูลงานได้", isUpdate)
		return
	}

	allJobs := store.GetRecentJobs(100) // Get up to 100 jobs
	totalJobs := len(allJobs)

	if totalJobs == 0 {
		respondWithMessage(s, i, fmt.Sprintf("%s ยังไม่มีงาน %s Internship ในขณะนี้", emoji, strings.Title(category)), isUpdate)
		return
	}

	// Calculate pagination
	totalPages := (totalJobs + jobsPerPage - 1) / jobsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	startIdx := (page - 1) * jobsPerPage
	endIdx := startIdx + jobsPerPage
	if endIdx > totalJobs {
		endIdx = totalJobs
	}

	jobs := allJobs[startIdx:endIdx]

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s Internships", emoji, strings.Title(category)),
		Description: fmt.Sprintf("📄 **หน้า %d/%d** | แสดง %d-%d จาก %d ตำแหน่ง", page, totalPages, startIdx+1, endIdx, totalJobs),
		Color:       getCategoryColor(category),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	for idx, job := range jobs {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", startIdx+idx+1, truncate(job.Title, 50)),
			Value:  fmt.Sprintf("🏢 %s\n📍 %s\n[Apply Now](%s)", truncate(job.Company, 30), truncate(job.Location, 20), job.URL),
			Inline: false,
		})
	}

	// Create pagination buttons
	var components []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent

	// Previous button
	if page > 1 {
		buttons = append(buttons, discordgo.Button{
			Label:    "◀️ ก่อนหน้า",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("jobs:%s:%d", category, page-1),
		})
	}

	// Page indicator
	buttons = append(buttons, discordgo.Button{
		Label:    fmt.Sprintf("หน้า %d/%d", page, totalPages),
		Style:    discordgo.SecondaryButton,
		CustomID: "page_indicator",
		Disabled: true,
	})

	// Next button
	if page < totalPages {
		buttons = append(buttons, discordgo.Button{
			Label:    "ถัดไป ▶️",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("jobs:%s:%d", category, page+1),
		})
	}

	if len(buttons) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: buttons,
		})
	}

	// Send response
	responseType := discordgo.InteractionResponseChannelMessageWithSource
	if isUpdate {
		responseType = discordgo.InteractionResponseUpdateMessage
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})

	log.Printf("✅ Sent page %d (%d jobs) to Discord", page, len(jobs))
}

func handleStatsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	categories := []string{"backend", "frontend", "fullstack"}
	var fields []*discordgo.MessageEmbedField
	total := 0

	for _, category := range categories {
		storageFile := fmt.Sprintf("data/%s_jobs.json", category)
		store := storage.NewStorage(storageFile)
		store.Load()
		count := store.GetJobCount()
		total += count

		emoji := getCategoryEmoji(category)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", emoji, strings.Title(category)),
			Value:  fmt.Sprintf("%d ตำแหน่ง", count),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📊 สถิติ Internship Jobs",
		Description: fmt.Sprintf("รวมทั้งหมด **%d** ตำแหน่ง", total),
		Color:       0x00b900,
		Fields:      fields,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func getCategoryColor(category string) int {
	switch category {
	case "backend":
		return 0x3498db // Blue
	case "frontend":
		return 0x2ecc71 // Green
	case "fullstack":
		return 0xf1c40f // Yellow
	default:
		return 0x95a5a6 // Gray
	}
}

func getCategoryEmoji(category string) string {
	switch category {
	case "backend":
		return "🔵"
	case "frontend":
		return "🟢"
	case "fullstack":
		return "🟡"
	default:
		return "📋"
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func respondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string, isUpdate bool) {
	responseType := discordgo.InteractionResponseChannelMessageWithSource
	if isUpdate {
		responseType = discordgo.InteractionResponseUpdateMessage
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string, isUpdate bool) {
	respondWithMessage(s, i, "❌ "+msg, isUpdate)
}

// Helper for models import
var _ = models.Job{}
