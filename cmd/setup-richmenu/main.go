package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Rich Menu structure
type RichMenu struct {
	Size        Size   `json:"size"`
	Selected    bool   `json:"selected"`
	Name        string `json:"name"`
	ChatBarText string `json:"chatBarText"`
	Areas       []Area `json:"areas"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Area struct {
	Bounds Bounds `json:"bounds"`
	Action Action `json:"action"`
}

type Bounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Action struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found")
	}

	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	if channelToken == "" {
		log.Fatal("❌ LINE_CHANNEL_TOKEN not set in .env")
	}

	fmt.Println("🎨 LINE Rich Menu Setup")
	fmt.Println("=======================\n")

	// Step 1: Create Rich Menu
	richMenuID, err := createRichMenu(channelToken)
	if err != nil {
		log.Fatalf("❌ Failed to create Rich Menu: %v", err)
	}
	fmt.Printf("✅ Rich Menu created: %s\n\n", richMenuID)

	// Step 2: Instructions for uploading image
	fmt.Println("📸 Next Steps:")
	fmt.Println("1. Create a Rich Menu image (2500x1686 pixels)")
	fmt.Println("   - 3 sections: Backend (left), Frontend (middle), Fullstack (right)")
	fmt.Println("   - Save as 'richmenu.png'")
	fmt.Println("")
	fmt.Println("2. Upload image with this command:")
	fmt.Printf("   curl -X POST https://api-data.line.me/v2/bot/richmenu/%s/content \\\n", richMenuID)
	fmt.Printf("     -H 'Authorization: Bearer %s' \\\n", channelToken)
	fmt.Println("     -H 'Content-Type: image/png' \\")
	fmt.Println("     --data-binary @richmenu.png")
	fmt.Println("")
	fmt.Println("3. Set as default Rich Menu:")
	fmt.Printf("   curl -X POST https://api.line.me/v2/bot/user/all/richmenu/%s \\\n", richMenuID)
	fmt.Printf("     -H 'Authorization: Bearer %s'\n", channelToken)
	fmt.Println("")
	fmt.Println("✅ Done! Rich Menu will appear in LINE")
}

func createRichMenu(channelToken string) (string, error) {
	// Define Rich Menu with 3 sections
	richMenu := RichMenu{
		Size: Size{
			Width:  2500,
			Height: 1686,
		},
		Selected:    true,
		Name:        "Internship Categories",
		ChatBarText: "เลือกประเภทงาน",
		Areas: []Area{
			// Backend (Left)
			{
				Bounds: Bounds{X: 0, Y: 0, Width: 833, Height: 1686},
				Action: Action{Type: "message", Text: "🔵 Backend Internships"},
			},
			// Frontend (Middle)
			{
				Bounds: Bounds{X: 833, Y: 0, Width: 834, Height: 1686},
				Action: Action{Type: "message", Text: "🟢 Frontend Internships"},
			},
			// Fullstack (Right)
			{
				Bounds: Bounds{X: 1667, Y: 0, Width: 833, Height: 1686},
				Action: Action{Type: "message", Text: "🟡 Fullstack Internships"},
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(richMenu)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	url := "https://api.line.me/v2/bot/richmenu"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+channelToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		RichMenuID string `json:"richMenuId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.RichMenuID, nil
}
