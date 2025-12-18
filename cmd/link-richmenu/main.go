package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found")
	}

	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	userID := os.Getenv("LINE_USER_ID")

	if channelToken == "" {
		log.Fatal("❌ LINE_CHANNEL_TOKEN not set")
	}

	fmt.Println("🔧 LINE Rich Menu Manager")
	fmt.Println("=========================\n")

	// List all Rich Menus
	richMenus, err := listRichMenus(channelToken)
	if err != nil {
		log.Fatalf("❌ Failed to list Rich Menus: %v", err)
	}

	fmt.Printf("📋 Found %d Rich Menus:\n", len(richMenus))

	var targetMenuID string
	for i, menu := range richMenus {
		fmt.Printf("   [%d] %s - %s (selected: %v)\n", i+1, menu.RichMenuID, menu.Name, menu.Selected)
		if menu.Name == "Internship Categories" {
			targetMenuID = menu.RichMenuID
		}
	}

	if targetMenuID == "" {
		log.Fatal("❌ Rich Menu 'Internship Categories' not found!")
	}

	fmt.Printf("\n✅ Target Rich Menu: %s\n", targetMenuID)

	// Set as default Rich Menu for ALL users
	fmt.Println("\n🌐 Setting Rich Menu as default for ALL users...")
	if err := setDefaultRichMenu(channelToken, targetMenuID); err != nil {
		log.Printf("⚠️  Failed to set default: %v", err)

		// Fallback: try to link to specific user
		if userID != "" {
			fmt.Printf("📱 Trying to link to user: %s\n", userID)
			if err := linkRichMenuToUser(channelToken, userID, targetMenuID); err != nil {
				log.Fatalf("❌ Failed to link: %v", err)
			}
			fmt.Println("✅ Rich Menu linked to user successfully!")
		}
	} else {
		fmt.Println("✅ Rich Menu set as default for ALL users!")
	}

	fmt.Println("\n🎉 Done! New friends will see the Rich Menu automatically.")
}

type RichMenu struct {
	RichMenuID string `json:"richMenuId"`
	Name       string `json:"name"`
	Selected   bool   `json:"selected"`
}

type RichMenuList struct {
	RichMenus []RichMenu `json:"richmenus"`
}

func listRichMenus(token string) ([]RichMenu, error) {
	req, _ := http.NewRequest("GET", "https://api.line.me/v2/bot/richmenu/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result RichMenuList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.RichMenus, nil
}

func linkRichMenuToUser(token, userID, richMenuID string) error {
	url := fmt.Sprintf("https://api.line.me/v2/bot/user/%s/richmenu/%s", userID, richMenuID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setDefaultRichMenu(token, richMenuID string) error {
	url := fmt.Sprintf("https://api.line.me/v2/bot/user/all/richmenu/%s", richMenuID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
