package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/domodwyer/mailyak"
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
		Name     string           `json:"Name"`
		To       []string         `json:"To"`
		Settings scheduleSettings `json:"Settings"`
	} `json:"Schedule"`
}

type scheduleSettings struct {
	Day         []int    `json:"Day"`
	Weekday     []string `json:"Weekday"`
	WorkingTime string   `json:"WorkingTime"`
}

type table struct {
	Duration                        int64
	Calldate, Disposition, Src, Dst string
}

var sending = []string{}

func IsSend(ss scheduleSettings) bool {
	now := time.Now()
	bDay := false
	bWeekday := false
	bWorkingTime := false
	if ss.Day != nil {
		for _, d := range ss.Day {
			if d == now.Day() {
				bDay = true
				break
			}
		}
	} else {
		bDay = true
	}
	if ss.Weekday != nil {
		for _, w := range ss.Weekday {
			//println(now.Weekday().String())
			if w == now.Weekday().String() {
				//println("args ...Type")
				bWeekday = true
				break
			}
		}
	} else {
		bWeekday = true
	}
	if ss.WorkingTime != "" {
		wt := ss.WorkingTime

		wts := strings.Split(wt, "-")
		fmt.Println(wts)
		b, err := time.Parse("15:04", wts[0])
		if err != nil {
			fmt.Println("err1: ", err)
			return false
		}
		e, err := time.Parse("15:04", wts[1])
		if err != nil {
			fmt.Println("err2: ", err)

			return false
		}
		if b.Unix() < e.Unix() {
			if now.Hour() >= b.Hour() && now.Hour() <= e.Hour() {
				if b.Minute() > 0 && b.Hour() == now.Hour() && b.Minute() < now.Minute() {
					bWorkingTime = true
				}
				if e.Minute() > 0 && e.Hour() == now.Hour() && e.Minute() > now.Minute() {
					bWorkingTime = true
				}
				if now.Hour() > b.Hour() && now.Hour() < e.Hour() {
					bWorkingTime = true
				}
			}
		} else {
			fmt.Printf("[%d:%d %d:%d]\n", b.Hour(), b.Minute(), e.Hour(), e.Minute())
			if b.Hour() <= now.Hour() || e.Hour() >= now.Hour() {
				if !bWorkingTime && b.Minute() > 0 && b.Minute() < now.Minute() {
					bWorkingTime = true
				}
				if !bWorkingTime && e.Minute() > now.Minute() {
					bWorkingTime = true
				}
				if b.Hour() < now.Hour() || e.Hour() > now.Hour() {
					bWorkingTime = true
				}
			}
		}
	} else {
		bWorkingTime = true
	}
	println(bDay, bWeekday, bWorkingTime)
	return bDay && bWeekday && bWorkingTime
}

func isSending(send string, sendings []string) bool {
	for _, s := range sendings {
		if s == send {
			return true
		}
	}
	return false
}

func sendEmail(EmailSettings emailSettings, data table) {
	t, err := template.ParseFiles(EmailSettings.Template)
	if err != nil {
		fmt.Println("Template: ", err)
		return
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		fmt.Println("Execute: ", err)
		return
	}
	sched := EmailSettings.Schedule
	//
	for _, setting := range sched {
		if IsSend(setting.Settings) {
			fmt.Println("setting.Name: ", setting.Name)
			mail := mailyak.New(EmailSettings.SMTPServer, nil)
			for _, to := range setting.To {
				if isSending(to, sending) {
					continue
				}
				mail.To(to)
				mail.From("gocdr_parcer@ath.ru")
				//mail.Subject("Call " + data.Src + " is NO ANSWER")
				mail.HTML().Set(buf.String())

				// input can be a bytes.Buffer, os.File, os.Stdin, etc.
				// call multiple times to attach multiple files
				logo, err := os.Open("bbb-logo.png")
				if err != nil {
					println(err)
					return
				}
				mail.AttachInline("image1", logo)
				telephone, err := os.Open("1425394958_telephone-64.png")
				if err != nil {
					println(err)
					return
				}
				mail.AttachInline("image2", telephone)

				if err := mail.Send(); err != nil {
					println(err.Error())
					//				panic(" 💣 ")
				}
				sending = append(sending, to)
			}
			//mail.Send()

		}
	}
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
		fmt.Println("err stmt1: ", err)
		return
	}
	defer stmt.Close()

	stmtdb, err := db.Prepare("SELECT descr, extension FROM queues_config where extension=?;")
	if err != nil {
		fmt.Println("err stmt2: ", err)
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
