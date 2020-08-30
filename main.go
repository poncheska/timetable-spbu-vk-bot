package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	vkapi "github.com/Dimonchik0036/vk-api"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
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
	regRegexp = regexp.MustCompile("^\\/reg https:\\/\\/timetable.spbu.ru\\/\\S+$")
	adminId   = int64(102727269)
)

const (
	UsersFilename = "users.json"
	ConnString    = "u7AxuyYlkB:HeXbIWd51j@tcp(remotemysql.com:3306)/u7AxuyYlkB"
)

func main() {
	//addr := os.Getenv("GRPC_ADDR")
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

	//SetJson(addr)
	users := GetUsers()
	//go JsonPusher(addr)

	go TTNotification(users, client)

	for update := range updates {
		if update.Message == nil || !update.IsNewMessage() || update.Message.Outbox() {
			continue
		}

		log.Printf("%s", update.Message.String())

		switch {
		case update.Message.Text == "/info":
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				">> Для регистрации введите \"/reg https://timetable.spbu.ru/...\"\n"+
					"(где ссылка указывает на расписание на текущую неделю\n"+
					"и не должна содержать дату на конце, пример ссылки:\n"+
					"\"https://timetable.spbu.ru/CHEM/StudentGroupEvents/Primary/276448\".)\n"+
					">> После регестрации используй \"/tt\" для получения расписания.\n"+
					">> Отмена регистрации \"/unreg\""))

		case StringStartsFrom(update.Message.Text, "/reg"):
			if !regRegexp.MatchString(update.Message.Text) {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Invalid link!"))
				continue
			}
			users.AddUser(update.Message.FromID, update.Message.Text[5:])
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Ты зарегестрирован!"))

		case update.Message.Text == "/users":
			if update.Message.FromID == adminId {
				bytes, err := json.Marshal(users)
				if err != nil {
					client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
						"Файл "+UsersFilename+" недоступен!!!\n"+err.Error()))
					continue
				}
				client.SendDoc(vkapi.NewDstFromUserID(update.Message.FromID), "users",
					vkapi.FileBytes{Bytes: bytes, Name: "users.txt"})
			} else {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не админ(("))
			}

		//case StringStartsFrom(update.Message.Text, "/load"):
		//	if update.Message.FromID == adminId {
		//		jsn := update.Message.Text[6:]
		//		err := ioutil.WriteFile(UsersFilename, []byte(jsn), os.FileMode(int(0777)))
		//		if err != nil {
		//			log.Println("load: " + err.Error())
		//		}
		//		users = GetUsers()
		//		client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
		//			"Юзеры загружены"))
		//	} else {
		//		client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
		//			"Ты не админ(("))
		//	}

		case update.Message.Text == "/unreg":
			err := users.DeleteUser(update.Message.FromID)
			if err != nil {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не был зарегестрирован!"))
			} else {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Регистрация отменена!"))
			}

		case update.Message.Text == "/tt":
			flag := true
			link := ""
			for _, u := range users.Users {
				log.Println(u)
				if u.ID == update.Message.FromID {
					log.Println(u)
					link = u.TTLink
					flag = false
					break
				}
			}
			if flag {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					"Ты не зарегистрирован"))
				continue
			}
			tt, err := parser.ParseTimetable(link)
			if err != nil {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
					err.Error()))
				continue
			}
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Расписание на неделю:\n"))
			for _, d := range tt.Days {
				strings := d.GetString()
				for _, str := range strings {
					client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), str))
				}
			}

		default:
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID),
				"Я тебя не понял или ты быканул!?(Напиши \"/info\")"))
		}

	}
}

func GetUsers() *TimetableUsers {
	users := &TimetableUsers{Users: make([]TimetableUser, 0, 0)}
	conn := DBConnection()
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM u7AxuyYlkB.Users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		tu := TimetableUser{}
		err = rows.Scan(&tu.ID, &tu.TTLink)
		if err != nil {
			panic(err)
		}
		users.Users = append(users.Users, tu)
	}
	log.Println("UPLOAD USERS")
	//bytes, err := ioutil.ReadFile(UsersFilename)
	//if err != nil {
	//	log.Println("GetUsers: " + err.Error())
	//	return users
	//}
	//err = json.Unmarshal(bytes, users)
	//if err != nil {
	//	log.Println("GetUsers: " + err.Error())
	//	return users
	//}
	return users
}

//func (tu *TimetableUsers) SetUsers() {
//	bytes, err := json.MarshalIndent(tu, "", "\t")
//	if err != nil {
//		log.Println("SetUsers: " + err.Error())
//	}
//	err = ioutil.WriteFile(UsersFilename, bytes, os.FileMode(int(0777)))
//	if err != nil {
//		log.Println("SetUsers: " + err.Error())
//	}
//}

func (tu *TimetableUsers) DeleteUser(id int64) error {
	conn := DBConnection()
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM u7AxuyYlkB.Users WHERE id = ?;", id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if count != 0 {
		_, err := conn.Exec("DELETE FROM u7AxuyYlkB.Users WHERE id = ?;", id)
		if err != nil {
			panic(err)
		}
		log.Println(fmt.Sprintf("DELETE: ID:%v", id))
	} else {
		log.Println(fmt.Sprintf("DELETE: NOTHING TO DELETE"))
		return fmt.Errorf("not registered")
	}

	for i, u := range tu.Users {
		if u.ID == id {
			tu.Users[i] = tu.Users[len(tu.Users)-1]
			tu.Users = tu.Users[:len(tu.Users)-1]
		}
	}
	return nil
}

func (tu *TimetableUsers) AddUser(id int64, link string) {
	conn := DBConnection()
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM u7AxuyYlkB.Users WHERE id = ?;", id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if count != 0 {
		_, err := conn.Exec("UPDATE u7AxuyYlkB.Users SET link = ? WHERE id = ?;", link, id)
		if err != nil {
			panic(err)
		}
		log.Println(fmt.Sprintf("UPDATE: ID:%v", id))
	} else {
		_, err := conn.Exec("INSERT INTO u7AxuyYlkB.Users (id, link) VALUES (?,?);", id, link)
		if err != nil {
			panic(err)
		}
		log.Println(fmt.Sprintf("CREATE: ID:%v", id))
	}

	for i, u := range tu.Users {
		if u.ID == id {
			tu.Users[i] = tu.Users[len(tu.Users)-1]
			tu.Users = tu.Users[:len(tu.Users)-1]
		}
	}
	tu.Users = append(tu.Users, TimetableUser{id, link})
	//tu.Mu.Lock()
	//tu.SetUsers()
	//tu.Mu.Unlock()
}

func TTNotification(users *TimetableUsers, client *vkapi.Client) {
	for {
		for _, u := range users.Users {
			tt, err := parser.ParseTimetable(u.TTLink)
			if err != nil {
				client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(u.ID),
					err.Error()))
				continue
			}
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(u.ID),
				"Расписание на неделю:\n"))
			for _, d := range tt.Days {
				strings := d.GetString()
				for _, str := range strings {
					client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(u.ID), str))
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

func StringStartsFrom(str, beg string) bool {
	if len(str) < len(beg) {
		return false
	} else {
		for i := 0; i < len(beg); i++ {
			if str[i] != beg[i] {
				return false
			}
		}
		return true
	}
}

func DBConnection() *sql.DB {
	conn, err := sql.Open("mysql", ConnString)
	if err != nil {
		panic(err)
	}
	return conn
}

// grpc functions (heroku doesn't support :'( )
//func SetJson(grpcAddress string) {
//	grpcConn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
//	if err != nil {
//		panic(err)
//	}
//
//	defer grpcConn.Close()
//
//	client := vault.NewJsonVaultClient(grpcConn)
//
//	res, _ := client.Get(context.Background(), &vault.Nothing{})
//	if res != nil {
//		ioutil.WriteFile(UsersFilename, res.Data, os.FileMode(int(0777)))
//	}
//}
//
//func PushJson(grpcAddress string) {
//	grpcConn, err := grpc.Dial(grpcAddress, grpc.WithInsecure())
//	if err != nil {
//		panic(err)
//	}
//
//	defer grpcConn.Close()
//
//	client := vault.NewJsonVaultClient(grpcConn)
//
//	bytes, _ := ioutil.ReadFile(UsersFilename)
//
//	client.Set(context.Background(), &vault.JsonBytes{Data: bytes})
//}
//
//func JsonPusher(grpcAddres string) {
//	for {
//		time.Sleep(5 * time.Minute)
//		log.Println("GRPC: json pushed")
//		PushJson(grpcAddres)
//	}
//}
