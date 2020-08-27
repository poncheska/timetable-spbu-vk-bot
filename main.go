package main

import "github.com/nikepan/govkbot/v2"
import "log"

var VKAdminID = 3759927
var VKToken = "efjr98j9fj8jf4j958jj4985jfj9joijerf0fj548jf94jfiroefije495jf48"

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
