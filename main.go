package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type configureStructure struct {
	ServerDB      string        `json:"ServerDB"`
	DatabaseCDR   string        `json:"DatabaseCDR"`
	Database      string        `json:"Database"`
	UserDB        string        `json:"UserDB"`
	PasswordDB    string        `json:"PasswordDB"`
	EmailSettings emailSettings `json:"EmailSettings"`
}

type emailSettings struct {
	Template   string `json:"Template"`
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
}

type table struct {
	Duration                        int64
	Calldate, Disposition, Src, Dst string
}

func sendEmail(EmailSettings emailSettings, data table) {
	t, err := template.ParseFiles(EmailSettings.Template)
	if err != nil {
		fmt.Println("Template: ", err)
		return
	}
	//var html io.Writer
	err = t.Execute(os.Stdout, data)
	if err != nil {
		fmt.Println("Execute: ", err)
		return
	}
	//fmt.Println(html)
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
	dbcdr, err := connectDB(config.ServerDB, config.UserDB, config.PasswordDB, config.DatabaseCDR)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	defer dbcdr.Close()

	db, err := connectDB(config.ServerDB, config.UserDB, config.PasswordDB, config.Database)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	defer db.Close()

	stmt, err := dbcdr.Prepare("select UNIX_TIMESTAMP(max(calldate)) - UNIX_TIMESTAMP(min(calldate)) as duration, calldate," +
		"disposition, src, dst " +
		"from cdr " +
		"where uniqueid=?;")
	if err != nil {
		fmt.Println("err stmt: ", err)
		return
	}
	defer stmt.Close()

	stmtdb, err := db.Prepare("SELECT descr, extension FROM queues_config where extension=?;")
	if err != nil {
		fmt.Println("err stmt: ", err)
		return
	}
	defer stmtdb.Close()

	time.Sleep(10)
	rows, err := stmt.Query(*uniqueid)
	if err != nil {
		fmt.Println("err Query: ", err)
		return
	}
	defer rows.Close()
	t := table{}

	for rows.Next() {
		rows.Scan(&t.Duration, &t.Calldate, &t.Disposition, &t.Src, &t.Dst)
		if t.Disposition == "NO ANSWER" && t.Dst[0] == '5' {
			rowsdb, err := stmtdb.Query(t.Dst)
			if err != nil {
				fmt.Println("Extension err: ", err)
				return
			}
			var descr, extension string
			for rowsdb.Next() {
				rowsdb.Scan(&descr, &extension)
			}
			t.Dst = descr
			sendEmail(config.EmailSettings, t)
		}
	}
	fmt.Println(*uniqueid)
	fmt.Println(*configure)
}
