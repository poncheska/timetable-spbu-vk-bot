package main

import (
	"fmt"
	vkapi "github.com/himidori/golang-vk-api"
)

var VKAdminID = 102727269
var VKToken = "b65675104740f1079ec2d76cfc1c36606e3d489defc69e3fc8f23db340eba4a8c2cbacef0c903a6309b69"


func main() {
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
