package main

import (
	"fmt"
	vkapi "github.com/himidori/golang-vk-api"
	"os"
)

//var VKAdminID = 102727269


func main() {
	VKToken := os.Getenv("BOT_TOKEN")
	client, err := vkapi.NewVKClientWithToken(VKToken, nil, false)
	if err != nil {
		panic(err)
	}
	// listening received messages
	client.AddLongpollCallback("msgin", func(m *vkapi.LongPollMessage) {
		fmt.Printf("new message received from uid %d\n", m.UserID)
	})

	client.AddLongpollCallback("msgout", func(m *vkapi.LongPollMessage) {
		fmt.Printf("sent message to uid %d\n", m.UserID)
	})

	client.ListenLongPollServer()
}
