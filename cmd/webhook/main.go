package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/annop07/internship-alert-bot/pkg/webhook"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	fmt.Println("🤖 LINE Webhook Server")
	fmt.Println("======================\n")

	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")

	if channelSecret == "" || channelToken == "" {
		log.Fatal("❌ LINE_CHANNEL_SECRET or LINE_CHANNEL_TOKEN not set")
	}

	handler, err := webhook.NewHandler(channelSecret, channelToken)
	if err != nil {
		log.Fatalf("❌ Failed to create webhook handler: %v", err)
	}

	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		log.Fatalf("❌ Failed to create LINE bot: %v", err)
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		events, err := bot.ParseRequest(r)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
				log.Println("❌ Invalid signature")
			} else {
				w.WriteHeader(500)
				log.Printf("❌ Parse error: %v", err)
			}
			return
		}

		if err := handler.HandleEvents(events); err != nil {
			log.Printf("❌ Error handling events: %v", err)
		}

		w.WriteHeader(200)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ Webhook server starting on port %s", port)
	log.Printf("📡 Webhook endpoint: http://localhost:%s/webhook", port)
	log.Println("🛑 Press Ctrl+C to stop\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
