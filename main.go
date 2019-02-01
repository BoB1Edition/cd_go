package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type configureStructure struct {
	ServerDB      string `json:"ServerDB"`
	Database      string `json:"Database"`
	UserDB        string `json:"UserDB"`
	PasswordDB    string `json:"PasswordDB"`
	EmailSettings struct {
		SMTPServer string `json:"SMTPServer"`
		From       string `json:"From"`
		Schedule   []struct {
			Name     string `json:"Name"`
			Settings struct {
				Day         []int    `json:"Day"`
				Weekday     []string `json:"Weekday"`
				WorkingTime string   `json:"WorkingTime"`
				To          []string `json:"To"`
			} `json:"Settings"`
		} `json:"Schedule"`
	} `json:"EmailSettings"`
}

type table struct {
	duration                        int64
	calldate, disposition, src, dst string
}

func connectDB(Server, User, Password, Databse string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", User, Password, Server, Databse)
	fmt.Println("connectionString: ", connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	configure := flag.String("configure", "config.json", "config file")
	flag.String("s", "fake param", "fake param, not used")
	uniqueid := flag.Float64("u", 0.0, "uniqueid")
	flag.Parse()
	if *uniqueid == 0 {
		fmt.Println("u requered param")
		return
	}
	config := configureStructure{}
	file, err := ioutil.ReadFile(*configure)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &config)
	db, err := connectDB(config.ServerDB, config.UserDB, config.PasswordDB, config.Database)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	defer db.Close()
	q := fmt.Sprintf("select UNIX_TIMESTAMP(max(calldate)) - UNIX_TIMESTAMP(min(calldate)) as duration, calldate,"+
		"disposition, src, dst"+
		"from cdr"+
		"where uniqueid='%f';", *uniqueid)
	fmt.Println(q)
	rows, err := db.Query("select UNIX_TIMESTAMP(max(calldate)) - UNIX_TIMESTAMP(min(calldate)) as duration, calldate,"+
		"disposition, src, dst"+
		"from cdr"+
		"where uniqueid='$1';", *uniqueid)

	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	defer rows.Close()
	t := table{}
	for rows.Next() {
		rows.Scan(&t)
		fmt.Println("t: ", t)
	}
	fmt.Println(*uniqueid)
	fmt.Println(*configure)
}
