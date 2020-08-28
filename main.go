package main

import (
	"encoding/json"
	"fmt"
	vkapi "github.com/Dimonchik0036/vk-api"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"
)

type TimetableUser struct {
	ID     int64  `json:"id"`
	TTLink string `json:"link"`
}

type TimetableUsers struct {
	Users []TimetableUser `json:"users"`
	Mu    sync.Mutex       `json:"-"`
}

var (
	regRegexp = regexp.MustCompile("^\\/reg https:\\/\\/timetable.spbu.ru\\/\\S+$")
	adminId = int64(102727269)
	usersFilename = "users.json"
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

	users := GetUsers()

	for update := range updates {
		if update.Message == nil || !update.IsNewMessage() || update.Message.Outbox() {
			continue
		}

		log.Printf("%s", update.Message.String())


		switch {
		case update.Message.Text == "/info":
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Для регистрации введите \"/reg https://timetable.spbu.ru/...\"\n"+
					"(где ссылка указывает на расписание на текущую неделю\n"+
					"и не должна содержать дату на конце, пример ссылки:\n"+
					"\"https://timetable.spbu.ru/CHEM/StudentGroupEvents/Primary/276448\")"))


		case update.Message.Text[:4] == "/reg":
			if !regRegexp.MatchString(update.Message.Text) {
				users.AddUser(update.Message.FromID, update.Message.Text[5:])
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Invalid link!"))
				continue
			}
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Alright!"))


		case update.Message.Text == "/users":
			if update.Message.FromID == adminId{
				bytes, err := ioutil.ReadFile(usersFilename)
				if err != nil{
					client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
						"Файл "+usersFilename+" недоступен!!!"))
				}
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					fmt.Sprintf("json: \n%v\n struct: \n%v\n", string(bytes), users.Users)))
			}else{
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не админ("))
			}


		default:
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Я тебя не понял или ты быканул!?"))
		}

	}
}

func GetUsers() *TimetableUsers {
	users := &TimetableUsers{}
	bytes, err := ioutil.ReadFile(usersFilename)
	if err != nil {
		file, err := ioutil.TempFile("", usersFilename)
		if err != nil {
			panic(err)
		}
		file.Close()
		return users
	}
	err = json.Unmarshal(bytes, users)
	if err != nil {
		panic(err)
	}
	return users
}

func (tu *TimetableUsers) SetUsers() {
	bytes, _ := json.MarshalIndent(tu,"","\t")
	ioutil.WriteFile(usersFilename, bytes, os.FileMode(int(0777)))
}

func (tu *TimetableUsers) AddUser(id int64, link string) {
	for i, u := range tu.Users{
		if u.ID == id{
			tu.Users[i] = tu.Users[len(tu.Users) - 1]
			tu.Users = tu.Users[:len(tu.Users) - 1]
		}
	}
	tu.Mu.Lock()
	tu.Users = append(tu.Users, TimetableUser{id, link})
	tu.SetUsers()
	tu.Mu.Unlock()
}
