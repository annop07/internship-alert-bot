package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/annop07/internship-alert-bot/pkg/models"
)

type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type DiscordWebhook struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

func (d *DiscordNotifier) SendJobAlert(job *models.Job) error {
	embed := d.createJobEmbed(job)

	webhook := DiscordWebhook{
		Username:  "Internship Alert Bot",
		AvatarURL: "https://cdn-icons-png.flaticon.com/512/3135/3135715.png",
		Content:   "🔥 **NEW INTERNSHIP FOUND!**",
		Embeds:    []DiscordEmbed{embed},
	}

	return d.sendWebhook(webhook)
}

func (d *DiscordNotifier) SendMultipleJobsAlert(jobs []*models.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	const maxEmbedsPerMessage = 10

	for i := 0; i < len(jobs); i += maxEmbedsPerMessage {
		end := i + maxEmbedsPerMessage
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		embeds := make([]DiscordEmbed, 0, len(batch))

		for _, job := range batch {
			embeds = append(embeds, d.createJobEmbed(job))
		}

		content := fmt.Sprintf("🔥 **Found %d NEW INTERNSHIP%s!** (Showing %d-%d)",
			len(jobs),
			pluralize(len(jobs)),
			i+1,
			end)

		webhook := DiscordWebhook{
			Username:  "Internship Alert Bot",
			AvatarURL: "https://cdn-icons-png.flaticon.com/512/3135/3135715.png",
			Content:   content,
			Embeds:    embeds,
		}

		if err := d.sendWebhook(webhook); err != nil {
			return err
		}

		if end < len(jobs) {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (d *DiscordNotifier) SendSummary(totalJobs, newJobs int) error {
	var color int
	var title string
	var description string

	if newJobs > 0 {
		color = 0x2ecc71 // Green
		title = fmt.Sprintf("✅ Found %d New Job%s!", newJobs, pluralize(newJobs))
		description = fmt.Sprintf("Total jobs tracked: **%d**\nNew jobs found: **%d**", totalJobs, newJobs)
	} else {
		color = 0x95a5a6 // Gray
		title = "✅ Scan Complete"
		description = fmt.Sprintf("Total jobs tracked: **%d**\nNo new jobs found this time.", totalJobs)
	}

	embed := DiscordEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Footer: &DiscordEmbedFooter{
			Text: "Internship Alert Bot",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	webhook := DiscordWebhook{
		Username:  "Internship Alert Bot",
		AvatarURL: "https://cdn-icons-png.flaticon.com/512/3135/3135715.png",
		Embeds:    []DiscordEmbed{embed},
	}

	return d.sendWebhook(webhook)
}

func (d *DiscordNotifier) createJobEmbed(job *models.Job) DiscordEmbed {

	description := job.Description
	if len(description) > 300 {
		description = description[:297] + "..."
	}
	if description == "" {
		description = "No description available"
	}

	formattedDesc := fmt.Sprintf("> %s\n\n👉 **[Click here to view details & apply](%s)**", description, job.URL)

	fields := []DiscordEmbedField{
		{
			Name:   "🏢 Company",
			Value:  fmt.Sprintf("**%s**", job.Company),
			Inline: true,
		},
		{
			Name:   "📍 Location",
			Value:  job.Location,
			Inline: true,
		},
	}

	if job.PostedDate != "" {
		fields = append(fields, DiscordEmbedField{
			Name:   "⏰ Posted",
			Value:  job.PostedDate,
			Inline: true,
		})
	}

	return DiscordEmbed{
		Title:       "⚡ " + job.Title,
		Description: formattedDesc,
		URL:         job.URL,
		Color:       0x00b0f4,
		Fields:      fields,
		Footer: &DiscordEmbedFooter{
			Text: fmt.Sprintf("Job ID: %s • %s", job.ID, time.Now().Format("15:04")),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (d *DiscordNotifier) sendWebhook(webhook DiscordWebhook) error {
	payload, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (d *DiscordNotifier) TestConnection() error {
	log.Println("📱 Testing Discord webhook...")

	embed := DiscordEmbed{
		Title:       "✅ Connection Test Successful!",
		Description: "Your Discord webhook is working correctly. You will receive notifications here when new internships are found.",
		Color:       0x2ecc71, // Green
		Footer: &DiscordEmbedFooter{
			Text: "Internship Alert Bot - Test Message",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	webhook := DiscordWebhook{
		Username:  "Internship Alert Bot",
		AvatarURL: "https://cdn-icons-png.flaticon.com/512/3135/3135715.png",
		Content:   "🤖 **Bot is now online!**",
		Embeds:    []DiscordEmbed{embed},
	}

	if err := d.sendWebhook(webhook); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	log.Println("✅ Discord webhook test successful!")
	return nil
}

func pluralize(count int) string {
	if count > 1 {
		return "S"
	}
	return ""
}
