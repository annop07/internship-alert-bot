package webhook

import (
	"fmt"
	"log"
	"strings"

	"github.com/annop07/internship-alert-bot/pkg/models"
	"github.com/annop07/internship-alert-bot/pkg/storage"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// Handler handles LINE webhook events
type Handler struct {
	bot *linebot.Client
}

// NewHandler creates a new webhook handler
func NewHandler(channelSecret, channelToken string) (*Handler, error) {
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create LINE bot: %w", err)
	}

	return &Handler{bot: bot}, nil
}

// HandleEvents processes LINE webhook events
func (h *Handler) HandleEvents(events []*linebot.Event) error {
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := h.handleTextMessage(event.ReplyToken, message.Text, event.Source.UserID); err != nil {
					log.Printf("Error handling text message: %v", err)
				}
			}
		}
	}
	return nil
}

// handleTextMessage processes text messages from Rich Menu buttons
func (h *Handler) handleTextMessage(replyToken, text, userID string) error {
	log.Printf("Received message: %s from user: %s", text, userID)

	// Determine category from message text
	var category string
	var emoji string

	switch {
	case strings.Contains(text, "Backend"):
		category = "backend"
		emoji = "🔵"
	case strings.Contains(text, "Frontend"):
		category = "frontend"
		emoji = "🟢"
	case strings.Contains(text, "Fullstack"):
		category = "fullstack"
		emoji = "🟡"
	default:
		// Not a Rich Menu button, ignore
		return nil
	}

	// Load jobs from storage
	storageFile := fmt.Sprintf("data/%s_jobs.json", category)
	store := storage.NewStorage(storageFile)

	if err := store.Load(); err != nil {
		log.Printf("Failed to load %s jobs: %v", category, err)
		return h.replyError(replyToken, "ขออภัยครับ ไม่สามารถดึงข้อมูลงานได้ในขณะนี้")
	}

	// Get recent jobs
	jobs := store.GetRecentJobs(10) // Get 10 most recent jobs

	if len(jobs) == 0 {
		message := fmt.Sprintf("%s ยังไม่มีงาน %s Internship ในขณะนี้ครับ", emoji, strings.Title(category))
		return h.replyText(replyToken, message)
	}

	// Send jobs as Flex Messages
	log.Printf("Sending %d %s jobs to user", len(jobs), category)
	return h.sendJobsFlexMessage(replyToken, jobs, category, emoji)
}

// sendJobsFlexMessage sends jobs as LINE Flex Messages
func (h *Handler) sendJobsFlexMessage(replyToken string, jobs []*models.Job, category, emoji string) error {
	// Create bubbles for each job
	var bubbles []*linebot.BubbleContainer

	// Limit to 10 jobs for carousel (LINE limit)
	count := len(jobs)
	if count > 10 {
		count = 10
	}

	for i := 0; i < count; i++ {
		bubble := h.createJobBubble(jobs[i])
		bubbles = append(bubbles, bubble)
	}

	// Create carousel
	carousel := &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbles,
	}

	// Create header message
	headerText := fmt.Sprintf("%s %s Internships\nพบ %d ตำแหน่ง", emoji, strings.Title(category), len(jobs))

	messages := []linebot.SendingMessage{
		linebot.NewTextMessage(headerText),
		linebot.NewFlexMessage("Job Listings", carousel),
	}

	// Reply to user
	if _, err := h.bot.ReplyMessage(replyToken, messages...).Do(); err != nil {
		return fmt.Errorf("failed to send flex messages: %w", err)
	}

	log.Printf("✅ Sent %d job(s) to user", len(jobs))
	return nil
}

// createJobBubble creates a Flex Bubble for a job
func (h *Handler) createJobBubble(job *models.Job) *linebot.BubbleContainer {
	return &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeMega,
		Header: &linebot.BoxComponent{
			Type:            linebot.FlexComponentTypeBox,
			Layout:          linebot.FlexBoxLayoutTypeVertical,
			BackgroundColor: "#00b900",
			PaddingAll:      linebot.FlexComponentPaddingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   "🆕 NEW INTERNSHIP",
					Weight: linebot.FlexTextWeightTypeBold,
					Color:  "#ffffff",
					Size:   linebot.FlexTextSizeTypeSm,
				},
			},
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   job.Title,
					Weight: linebot.FlexTextWeightTypeBold,
					Size:   linebot.FlexTextSizeTypeLg,
					Wrap:   true,
				},
				&linebot.BoxComponent{
					Type:    linebot.FlexComponentTypeBox,
					Layout:  linebot.FlexBoxLayoutTypeVertical,
					Margin:  linebot.FlexComponentMarginTypeMd,
					Spacing: linebot.FlexComponentSpacingTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type: linebot.FlexComponentTypeText,
									Text: "🏢",
									Flex: intPtr(0),
									Size: linebot.FlexTextSizeTypeSm,
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  job.Company,
									Wrap:  true,
									Size:  linebot.FlexTextSizeTypeSm,
									Color: "#666666",
								},
							},
						},
						&linebot.BoxComponent{
							Type:    linebot.FlexComponentTypeBox,
							Layout:  linebot.FlexBoxLayoutTypeBaseline,
							Spacing: linebot.FlexComponentSpacingTypeSm,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type: linebot.FlexComponentTypeText,
									Text: "📍",
									Flex: intPtr(0),
									Size: linebot.FlexTextSizeTypeSm,
								},
								&linebot.TextComponent{
									Type:  linebot.FlexComponentTypeText,
									Text:  job.Location,
									Wrap:  true,
									Size:  linebot.FlexTextSizeTypeSm,
									Color: "#666666",
								},
							},
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeSm,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Height: linebot.FlexButtonHeightTypeSm,
					Action: &linebot.URIAction{
						Label: "Apply Now",
						URI:   job.URL,
					},
				},
			},
		},
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}

// replyText sends a simple text reply
func (h *Handler) replyText(replyToken, text string) error {
	if _, err := h.bot.ReplyMessage(replyToken, linebot.NewTextMessage(text)).Do(); err != nil {
		return fmt.Errorf("failed to reply text: %w", err)
	}
	return nil
}

// replyError sends an error message
func (h *Handler) replyError(replyToken, errorMsg string) error {
	return h.replyText(replyToken, errorMsg)
}
