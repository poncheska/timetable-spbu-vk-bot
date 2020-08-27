package main

import (
	vkapi "github.com/Dimonchik0036/vk-api"
	"log"
	"os"
)

func main() {
	//client, err := vkapi.NewClientFromLogin("<username>", "<password>", vkapi.ScopeMessages)
	client, err := vkapi.NewClientFromToken(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	client.Log(true)

	if err := client.InitLongPoll(0, 2); err != nil {
		log.Panic(err)
	}

	updates, _, err := client.GetLPUpdatesChan(100, vkapi.LPConfig{25, vkapi.LPModeAttachments})
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil || !update.IsNewMessage() || update.Message.Outbox() {
			continue
		}

		log.Printf("%s", update.Message.String())
		switch update.Message.Text {
		case "/start":
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Welcome to the club buddy!"))
		default:
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Я тебя не понял или ты быканул!?"))
		}

	}
}
