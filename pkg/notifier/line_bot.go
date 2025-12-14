package notifier

import (
	"fmt"
	"log"

	"github.com/annop07/internship-alert-bot/pkg/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// LineBotNotifier handles LINE Messaging API notifications
type LineBotNotifier struct {
	client *linebot.Client
	userID string
}

// NewLineBotNotifier creates a new LINE Bot notifier
func NewLineBotNotifier(channelSecret, channelToken, userID string) (*LineBotNotifier, error) {
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		return nil, err
	}
	return &LineBotNotifier{
		client: bot,
		userID: userID,
	}, nil
}

// SendJobAlert sends a notification for a single job
func (l *LineBotNotifier) SendJobAlert(job *models.Job) error {
	flexContainer := l.createJobFlexMessage(job)
	_, err := l.client.PushMessage(l.userID, linebot.NewFlexMessage("New Internship Alert!", flexContainer)).Do()
	return err
}

// SendMultipleJobsAlert sends a notification for multiple jobs using Carousel
func (l *LineBotNotifier) SendMultipleJobsAlert(jobs []*models.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	// LINE limits Carousel to 12 bubbles (previously 10, but 12 is safe)
	// We'll batch them in groups of 10
	batchSize := 10
	for i := 0; i < len(jobs); i += batchSize {
		end := i + batchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		var bubbles []*linebot.BubbleContainer

		for _, job := range batch {
			bubbles = append(bubbles, l.createJobBubble(job))
		}

		carousel := &linebot.CarouselContainer{
			Type:     linebot.FlexContainerTypeCarousel,
			Contents: bubbles,
		}

		altText := fmt.Sprintf("Found %d new internships!", len(batch))
		_, err := l.client.PushMessage(l.userID, linebot.NewFlexMessage(altText, carousel)).Do()
		if err != nil {
			log.Printf("⚠️ Failed to send LINE batch: %v", err)
			return err
		}
	}

	return nil
}

// SendSummary sends a text summary
func (l *LineBotNotifier) SendSummary(totalJobs, newJobs int) error {
	var message string
	if newJobs > 0 {
		message = fmt.Sprintf("✅ Scan Complete\nTotal Jobs: %d\n🔥 New Jobs: %d", totalJobs, newJobs)
	} else {
		message = fmt.Sprintf("✅ Scan Complete\nTotal Jobs: %d\nNo new jobs found.", totalJobs)
	}

	_, err := l.client.PushMessage(l.userID, linebot.NewTextMessage(message)).Do()
	return err
}

// TestConnection sends a test message
func (l *LineBotNotifier) TestConnection() error {
	log.Println("📱 Testing LINE Bot...")
	_, err := l.client.PushMessage(l.userID, linebot.NewTextMessage("✅ Internship Alert Bot (Messaging API) is online!")).Do()
	if err == nil {
		log.Println("✅ LINE Bot test successful!")
	}
	return err
}

// createJobFlexMessage creates a Flex Bubble for a single job
func (l *LineBotNotifier) createJobFlexMessage(job *models.Job) linebot.FlexContainer {
	return l.createJobBubble(job)
}

// createJobBubble creates the Bubble container
func (l *LineBotNotifier) createJobBubble(job *models.Job) *linebot.BubbleContainer {
	return &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeMega,
		Header: &linebot.BoxComponent{
			Type:            linebot.FlexComponentTypeBox,
			Layout:          linebot.FlexBoxLayoutTypeVertical,
			BackgroundColor: "#00b900", // LINE Green
			PaddingAll:      linebot.FlexComponentPaddingTypeMd,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   "NEW INTERNSHIP",
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
					Size:   linebot.FlexTextSizeTypeXl,
					Wrap:   true,
				},
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   job.Company,
					Size:   linebot.FlexTextSizeTypeMd,
					Color:  "#666666",
					Margin: linebot.FlexComponentMarginTypeSm,
					Wrap:   true,
				},
				&linebot.BoxComponent{
					Type:   linebot.FlexComponentTypeBox,
					Layout: linebot.FlexBoxLayoutTypeBaseline,
					Margin: linebot.FlexComponentMarginTypeMd,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "📍",
							Flex:  linebot.IntPtr(1),
							Size:  linebot.FlexTextSizeTypeSm,
							Color: "#aaaaaa",
						},
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  job.Location,
							Flex:  linebot.IntPtr(5),
							Size:  linebot.FlexTextSizeTypeSm,
							Color: "#666666",
							Wrap:  true,
						},
					},
				},
				&linebot.BoxComponent{
					Type:   linebot.FlexComponentTypeBox,
					Layout: linebot.FlexBoxLayoutTypeBaseline,
					Margin: linebot.FlexComponentMarginTypeSm,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  "⏰",
							Flex:  linebot.IntPtr(1),
							Size:  linebot.FlexTextSizeTypeSm,
							Color: "#aaaaaa",
						},
						&linebot.TextComponent{
							Type:  linebot.FlexComponentTypeText,
							Text:  job.PostedDate,
							Flex:  linebot.IntPtr(5),
							Size:  linebot.FlexTextSizeTypeSm,
							Color: "#666666",
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
