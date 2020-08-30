package createdb

import (
	"database/sql"
	_ "github.com/lib/pq"
	"io/ioutil"
)

var (
	dbAddr = "postgres://dwonflbcazjoon:fc674b2450912b60f6e3c96782f0bf50808b0d28c9df25496588a48eae08a7f4@ec2-54-247-103-43.eu-west-1.compute.amazonaws.com:5432/d4ie9jc8suj2i0"
)

func CreateDB() {
	db, err := sql.Open("postgres", dbAddr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query, _ := ioutil.ReadFile("users.sql")
	_, err = db.Exec(string(query))
	if err != nil{
		panic(err)
	}
}
