package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/leirbagxis/musicytbot/client/youtube"
)

func parseISODuration(iso string) string {
	iso = strings.ReplaceAll(iso, "PT", "")
	iso = strings.ReplaceAll(iso, "H", "h")
	iso = strings.ReplaceAll(iso, "M", "m")
	iso = strings.ReplaceAll(iso, "S", "s")
	dur, err := time.ParseDuration(strings.ToLower(iso))
	if err != nil {
		return iso
	}

	totalSeconds := int(dur.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Erro ao buscar detalhes do vídeo")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}

			var buttons [][]tgbotapi.InlineKeyboardButton
			for i, music := range youtubeDetails {
				if i >= 10 {
					break
				}

				callbackData := fmt.Sprintf("track_id:%s:yt", music.VideoID)
				formattedDuration := parseISODuration(music.Duration)
				buttonText := fmt.Sprintf("• %s • %s", formattedDuration, music.Title)
				button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
				buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
			}

			replyMarkup := tgbotapi.NewInlineKeyboardMarkup(buttons...)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ㅤ")
			msg.ReplyMarkup = replyMarkup
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}

		if update.CallbackQuery != nil {
			callbackData := update.CallbackQuery.Data
			parts := strings.Split(callbackData, ":")

			if len(parts) == 3 && parts[0] == "track_id" && parts[2] == "yt" {
				videoID := parts[1]
				videoURL := fmt.Sprintf("https://youtu.be/%s", videoID)

				// Responde o callback para tirar o "loading" no Telegram
				answerCallback := tgbotapi.NewCallback(update.CallbackQuery.ID, "⏳ Baixando o áudio...")
				bot.Request(answerCallback)

				// Busca detalhes da música para usar nos metadados
				youtubeDetails, err := ytClient.GetMusicDetails(videoID)
				if err != nil {
					log.Printf("Error fetching video details for metadata: %v", err)
				}

				// Define nome do arquivo temporário
				outputFile := fmt.Sprintf("%s.m4a", videoID)

				// Executa yt-dlp para baixar o áudio
				cmd := exec.Command("yt-dlp", "-x", "--audio-format", "m4a", "-o", outputFile, videoURL)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err = cmd.Run()
				if err != nil {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Erro ao baixar o áudio.")
					bot.Send(msg)
					continue
				}

				// Abre o arquivo e envia como áudio
				audioFile, err := os.Open(outputFile)
				if err != nil {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Erro ao abrir o áudio baixado.")
					bot.Send(msg)
					continue
				}
				defer audioFile.Close()
				defer os.Remove(outputFile) // Limpa o arquivo depois

				audioMsg := tgbotapi.NewAudio(update.CallbackQuery.Message.Chat.ID, tgbotapi.FileReader{
					Name:   outputFile,
					Reader: audioFile,
				})

				// Define metadados para aparecer no player do Telegram
				if len(youtubeDetails) > 0 {
					audioMsg.Title = youtubeDetails[0].Title
					audioMsg.Performer = youtubeDetails[0].Title
					audioMsg.Caption = fmt.Sprintf("🎵 @%s | <a href=\"https://song.link/y/%s\">Info</a>", bot.Self.UserName, youtubeDetails[0].VideoID)
					audioMsg.ParseMode = "HTML"
				} else {
					audioMsg.Caption = fmt.Sprintf("🎵 Aqui está a música: %s", videoURL)
				}

				_, err = bot.Send(audioMsg)
				if err != nil {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("❌ Erro ao enviar o áudio: %s", err.Error()))
					bot.Send(msg)
				}
			}
		}
	}
}
