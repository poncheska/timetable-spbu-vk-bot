package main

import (
	vkapi "github.com/Dimonchik0036/vk-api"
	"log"
	"os"
	"regexp"
)

type TimetableUser struct {
	ID     int64 `json:"id"`
	TTLink string `json:"link"`
}

type TimetableUsers struct {
	Users []TimetableUsers `json:"users"`
}

var (
	regRegexp = regexp.MustCompile("\\/reg https:\\/\\/timetable.spbu.ru\\/\\S+")
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
		switch {
		case update.Message.Text[:3] == "/reg":
			if !regRegexp.MatchString(update.Message.Text){
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Invalid link!"))
				continue
			}
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Alright!"))
		default:
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Я тебя не понял или ты быканул!?"))
			//file, err := ioutil.ReadFile("kaban.jpg")
			//if err != nil {
			//	log.Panic(err)
			//}
			//client.SendPhoto(vkapi.NewDstFromUserID(update.Message.FromID),
			//	vkapi.FileBytes{Bytes: file, Name: "kaban.jpg"})
		}

	}
}
