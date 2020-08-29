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
	"vk-timetable-bot/parser"
)

type TimetableUser struct {
	ID     int64  `json:"id"`
	TTLink string `json:"link"`
}

type TimetableUsers struct {
	Users []TimetableUser `json:"users"`
	Mu    sync.Mutex      `json:"-"`
}

var (
	regRegexp     = regexp.MustCompile("^\\/reg https:\\/\\/timetable.spbu.ru\\/\\S+$")
	adminId       = int64(102727269)
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
					"\"https://timetable.spbu.ru/CHEM/StudentGroupEvents/Primary/276448\".)\n" +
				"После регестрации используй \"/tt\" для получения расписания."))

		case StringStartsFrom(update.Message.Text,"/reg"):
			if !regRegexp.MatchString(update.Message.Text) {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Invalid link!"))
				continue
			}
			users.AddUser(update.Message.FromID, update.Message.Text[5:])
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Alright!"))

		case update.Message.Text == "/users":
			if update.Message.FromID == adminId {
				bytes, err := ioutil.ReadFile(usersFilename)
				if err != nil {
					client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
						"Файл "+usersFilename+" недоступен!!!\n"+err.Error()))
					continue
				}
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					fmt.Sprintf("json: \n%v\n struct: \n%v\n", string(bytes), users.Users)))
			} else {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не админ("))
			}

		case StringStartsFrom(update.Message.Text,"/load"):
			if update.Message.FromID == adminId {
				jsn := update.Message.Text[6:]
				err := ioutil.WriteFile(usersFilename, []byte(jsn), os.FileMode(int(0777)))
				if err != nil {
					log.Println("load: " + err.Error())
				}
				users = GetUsers()
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Юзеры загружены"))
			} else {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не админ("))
			}

		case update.Message.Text == "/tt":
			log.Println(cap(users.Users), len(users.Users))
			flag := true
			link := ""
			for _, u := range users.Users{
				log.Println(u)
				if u.ID == update.Message.FromID{
					log.Println(u)
					link = u.TTLink
					flag = false
					break
				}
			}
			if flag{
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не зарегистрирован"))
				continue
			}
			tt, err := parser.ParseTimetable(link)
			if err != nil || tt == nil{
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Что-то не так с твоей ссылкой зарегистрируйся заново"))
				continue
			}
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "tt.GetString()"))



		default:
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Я тебя не понял или ты быканул!?"))
		}

	}
}

func GetUsers() *TimetableUsers {
	users := &TimetableUsers{Users: make([]TimetableUser, 0, 0)}
	bytes, err := ioutil.ReadFile(usersFilename)
	if err != nil {
		log.Println("GetUsers: " + err.Error())
		return users
	}
	err = json.Unmarshal(bytes, users)
	if err != nil {
		log.Println("GetUsers: " + err.Error())
		return users
	}
	return users
}

func (tu *TimetableUsers) SetUsers() {
	bytes, err := json.MarshalIndent(tu, "", "\t")
	if err != nil {
		log.Println("SetUsers: " + err.Error())
	}
	err = ioutil.WriteFile(usersFilename, bytes, os.FileMode(int(0777)))
	if err != nil {
		log.Println("SetUsers: " + err.Error())
	}
}

func (tu *TimetableUsers) AddUser(id int64, link string) {
	for i, u := range tu.Users {
		if u.ID == id {
			tu.Users[i] = tu.Users[len(tu.Users)-1]
			tu.Users = tu.Users[:len(tu.Users)-1]
		}
	}
	tu.Mu.Lock()
	tu.Users = append(tu.Users, TimetableUser{id, link})
	tu.SetUsers()
	tu.Mu.Unlock()
}

func StringStartsFrom(str, beg string)bool{
	if len(str) < len(beg){
		return false
	}else {
		for i:=0;i<len(beg);i++{
			if str[i]!=beg[i]{
				return false
			}
		}
		return true
	}
}
