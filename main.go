package main

import "github.com/nikepan/govkbot/v2"
import "log"

var VKAdminID = 102727269
var VKToken = "b65675104740f1079ec2d76cfc1c36606e3d489defc69e3fc8f23db340eba4a8c2cbacef0c903a6309b69"

func helpHandler(m *govkbot.Message) (reply string) {
	return "help received"
}

func startHandler(m *govkbot.Message) (reply govkbot.Reply) {
	keyboard := govkbot.Keyboard{Buttons: make([][]govkbot.Button, 0)}
	button := govkbot.NewButton("/help", nil)
	row := make([]govkbot.Button, 0)
	row = append(row, button)
	keyboard.Buttons = append(keyboard.Buttons, row)

	return govkbot.Reply{Msg: "availableCommands", Keyboard: &keyboard}
}

func errorHandler(m *govkbot.Message, err error) {
	log.Fatal(err.Error())
}

func main() {
	//govkbot.HandleMessage("/", anyHandler)
	//govkbot.HandleMessage("/me", meHandler)
	govkbot.HandleMessage("/help", helpHandler)
	govkbot.HandleAdvancedMessage("/start", startHandler)

	//govkbot.HandleAction("chat_invite_user", inviteHandler)
	//govkbot.HandleAction("chat_kick_user", kickHandler)
	//govkbot.HandleAction("friend_add", addFriendHandler)
	//govkbot.HandleAction("friend_delete", deleteFriendHandler)

	govkbot.HandleError(errorHandler)

	govkbot.SetAutoFriend(true) // enable auto accept/delete friends

	govkbot.SetDebug(true) // log debug messages

	// Optional Direct VK API access
	govkbot.SetAPI(VKToken, "", "") // Need only before Listen, if you use direct API
	me, _ := govkbot.API.Me() // call API method
	log.Printf("current user: %+v\n", me.FullName())
	// Optional end

	govkbot.Listen(VKToken, "", "", VKAdminID)
}
