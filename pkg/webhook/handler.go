package webhook

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/annop07/internship-alert-bot/pkg/models"
	"github.com/annop07/internship-alert-bot/pkg/storage"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

const jobsPerPage = 10

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
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := h.handleTextMessage(event.ReplyToken, message.Text, event.Source.UserID); err != nil {
					log.Printf("Error handling text message: %v", err)
				}
			}
		case linebot.EventTypePostback:
			if err := h.handlePostback(event.ReplyToken, event.Postback.Data, event.Source.UserID); err != nil {
				log.Printf("Error handling postback: %v", err)
			}
		}
	}
	return nil
}

func (h *Handler) handlePostback(replyToken, data, userID string) error {
	log.Printf("Received postback: %s from user: %s", data, userID)

	parts := strings.Split(data, "&")
	params := make(map[string]string)
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	category := params["category"]
	pageStr := params["page"]
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	return h.sendJobsPage(replyToken, category, page)
}

// handleTextMessage processes text messages from Rich Menu buttons
func (h *Handler) handleTextMessage(replyToken, text, userID string) error {
	log.Printf("Received message: %s from user: %s", text, userID)

	// Determine category from message text
	var category string

	switch {
	case strings.Contains(text, "Backend"):
		category = "backend"
	case strings.Contains(text, "Frontend"):
		category = "frontend"
	case strings.Contains(text, "Fullstack"):
		category = "fullstack"
	default:
		return nil
	}

	return h.sendJobsPage(replyToken, category, 1)
}

// sendJobsPage sends a page of jobs with pagination
func (h *Handler) sendJobsPage(replyToken, category string, page int) error {
	emoji := getCategoryEmoji(category)

	// Load jobs from storage
	storageFile := fmt.Sprintf("data/%s_jobs.json", category)
	store := storage.NewStorage(storageFile)

	if err := store.Load(); err != nil {
		log.Printf("Failed to load %s jobs: %v", category, err)
		return h.replyError(replyToken, "ขออภัยครับ ไม่สามารถดึงข้อมูลงานได้ในขณะนี้")
	}

	allJobs := store.GetRecentJobs(100)
	totalJobs := len(allJobs)

	if totalJobs == 0 {
		message := fmt.Sprintf("%s ยังไม่มีงาน %s Internship ในขณะนี้ครับ", emoji, strings.Title(category))
		return h.replyText(replyToken, message)
	}

	// Calculate pagination
	startIdx := (page - 1) * jobsPerPage
	endIdx := startIdx + jobsPerPage
	if endIdx > totalJobs {
		endIdx = totalJobs
	}
	if startIdx >= totalJobs {
		return h.replyText(replyToken, "ไม่มีงานเพิ่มเติมแล้วครับ")
	}

	jobs := allJobs[startIdx:endIdx]
	hasMore := endIdx < totalJobs

	log.Printf("Sending page %d (%d-%d of %d) %s jobs to user", page, startIdx+1, endIdx, totalJobs, category)

	return h.sendJobsFlexMessage(replyToken, jobs, category, emoji, page, totalJobs, hasMore)
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

func (h *Handler) sendJobsFlexMessage(replyToken string, jobs []*models.Job, category, emoji string, page, totalJobs int, hasMore bool) error {
	var bubbles []*linebot.BubbleContainer

	for _, job := range jobs {
		bubble := h.createJobBubble(job)
		bubbles = append(bubbles, bubble)
	}

	if hasMore {
		nextPageBubble := h.createNextPageBubble(category, page+1, totalJobs)
		bubbles = append(bubbles, nextPageBubble)
	}

	carousel := &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbles,
	}

	startNum := (page-1)*jobsPerPage + 1
	endNum := startNum + len(jobs) - 1
	headerText := fmt.Sprintf("%s %s Internships\n📄 หน้า %d | แสดง %d-%d จาก %d ตำแหน่ง",
		emoji, strings.Title(category), page, startNum, endNum, totalJobs)

	messages := []linebot.SendingMessage{
		linebot.NewTextMessage(headerText),
		linebot.NewFlexMessage("Job Listings", carousel),
	}

	// Reply to user
	if _, err := h.bot.ReplyMessage(replyToken, messages...).Do(); err != nil {
		return fmt.Errorf("failed to send flex messages: %w", err)
	}

	log.Printf("✅ Sent page %d (%d jobs) to user", page, len(jobs))
	return nil
}

// createNextPageBubble creates a bubble for "View More" pagination
func (h *Handler) createNextPageBubble(category string, nextPage, totalJobs int) *linebot.BubbleContainer {
	emoji := getCategoryEmoji(category)
	remaining := totalJobs - (nextPage-1)*jobsPerPage
	if remaining > jobsPerPage {
		remaining = jobsPerPage
	}

	return &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeMega,
		Body: &linebot.BoxComponent{
			Type:            linebot.FlexComponentTypeBox,
			Layout:          linebot.FlexBoxLayoutTypeVertical,
			JustifyContent:  linebot.FlexComponentJustifyContentTypeCenter,
			AlignItems:      linebot.FlexComponentAlignItemsTypeCenter,
			BackgroundColor: "#f5f5f5",
			PaddingAll:      linebot.FlexComponentPaddingTypeXxl,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "📄",
					Size:  linebot.FlexTextSizeType3xl,
					Align: linebot.FlexComponentAlignTypeCenter,
				},
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   "ดูงานเพิ่มเติม",
					Weight: linebot.FlexTextWeightTypeBold,
					Size:   linebot.FlexTextSizeTypeLg,
					Align:  linebot.FlexComponentAlignTypeCenter,
					Margin: linebot.FlexComponentMarginTypeMd,
				},
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   fmt.Sprintf("%s ยังมีอีก %d+ ตำแหน่ง", emoji, remaining),
					Size:   linebot.FlexTextSizeTypeSm,
					Color:  "#666666",
					Align:  linebot.FlexComponentAlignTypeCenter,
					Margin: linebot.FlexComponentMarginTypeSm,
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
					Color:  "#00b900",
					Height: linebot.FlexButtonHeightTypeSm,
					Action: &linebot.PostbackAction{
						Label: fmt.Sprintf("ดูหน้า %d →", nextPage),
						Data:  fmt.Sprintf("category=%s&page=%d", category, nextPage),
					},
				},
			},
		},
	}
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
