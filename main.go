package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/leirbagxis/musicytbot/client/youtube"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	ytClient := youtube.New(os.Getenv("YTAPI_KEY"))

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			youtubeDetails, err := ytClient.GetMusicDetails(update.Message.Text)
			log.Printf("Error fetching video details: %v", err)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Erro r fetching video details")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}

			var reply string
			for _, detail := range youtubeDetails {
				reply += fmt.Sprintf("🆔 ID: %s\n📌 Titulo: %s\n📎 Link: https://www.youtube.com/watch?v=%s\n\n",
					detail.VideoID, detail.Title, detail.VideoID)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
