package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/snowzhop/go-s2t-bot/internal/voice"
	vosklocal "github.com/snowzhop/go-s2t-bot/internal/vosk/local"
	"github.com/snowzhop/go-s2t-bot/tools"
)

const (
	TGBOTAPI_ENV_VAR = "TGBOTAPI_TOKEN"
)

type UserData struct {
	Username string
	ID       int64
}

func main() {
	fmt.Println("go-s2t-bot V2")

	botToken := os.Getenv(TGBOTAPI_ENV_VAR)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Can't create new bot api: %v", err)
	}
	fmt.Println("Created new bot API")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 10

	botUpdates := bot.GetUpdatesChan(updateConfig)

	// u := url.URL{Scheme: "ws", Host: "localhost:2700", Path: ""}
	// log.Printf("Vosk-Server: %s", u.String())

	voskAdp, err := vosklocal.NewAdapter()
	if err != nil {
		log.Fatalf("Can't create adapter for local recognizer: %v", err)
	}

	voskReqIDtoChatID := make(map[uint64]*UserData)

	voskResults := voskAdp.ResultsChan()

	fmt.Println("Ready to recognize!")

	for {
		select {
		case update := <-botUpdates:
			if update.Message != nil {
				switch {
				case update.Message.Voice != nil:
					v := update.Message.Voice

					url, err := bot.GetFileDirectURL(v.FileID)
					if err != nil {
						log.Printf("[%s] error: %v", update.Message.From.UserName, err)
						break
					}

					voiceMsg, err := tools.DownloadData(url)
					if err != nil {
						log.Printf("Can't get voice message: %v", err)
						break
					}
					fmt.Printf("Raw audio len: %d\n", len(voiceMsg))

					wavVoice, err := voice.OpusToWav(voiceMsg, v.Duration)
					if err != nil {
						log.Printf("opus to wav coverting error: %v", err)
						break
					}
					fmt.Printf("Decoded len: %d\n", len(wavVoice))

					id := voskAdp.Recognize(wavVoice)
					voskReqIDtoChatID[id] = &UserData{
						Username: update.Message.From.UserName,
						ID:       update.Message.Chat.ID,
					}

				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, `¯\_(ツ)_/¯`)
					bot.Send(msg)
				}
			}
		case answer := <-voskResults:
			if answer.Error != nil {
				log.Printf("Recognizing error: %v", answer.Error)
				continue
			}

			user := voskReqIDtoChatID[answer.ID]

			log.Printf("for '%s': %s", user.Username, answer.Text)

			msg := tgbotapi.NewMessage(user.ID, answer.Text)
			bot.Send(msg)
		}
	}
}
