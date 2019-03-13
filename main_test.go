package main

import (
	"fmt"
	"testing"
	"time"
)

//import "main.go"

func TestIsSend(t *testing.T) {
	now := time.Now()
	s := scheduleSettings{}
	//s.Day = [1,2,3,4,5,6,7,8]
	//fmt.Printf("%d:00-%d:00", now.Hour()-1, now.Hour()+1)
	s.WorkingTime = fmt.Sprintf("%02d:00-%02d:00", now.Hour()-1, now.Hour()+1)
	b := IsSend(s)
	if !b {
		t.Errorf("Test 1 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:%02d-23:59", now.Hour(), now.Minute()-1)
	b = IsSend(s)
	if !b {
		t.Errorf("Test 2 false: %02s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("0:00-%02d:%02d", now.Hour(), now.Minute()+1)
	b = IsSend(s)
	if !b {
		t.Errorf("Test 3 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:00-%02d:00", now.Hour()-2, now.Hour()-1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 4 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:00-%02d:00", now.Hour()+1, now.Hour()+2)
	b = IsSend(s)
	if b {
		t.Errorf("Test 5 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:%02d-23:59", now.Hour(), now.Minute()+1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 6 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("0:00-%02d:%02d", now.Hour(), now.Minute()-1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 7 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:00-%02d:00", now.Hour()+1, now.Hour()-1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 8 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:00-%02d:00", now.Hour()-1, now.Hour()-3)
	b = IsSend(s)
	if !b {
		t.Errorf("Test 9 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:%02d-1:00", now.Hour(), now.Minute()-1)
	b = IsSend(s)
	if !b {
		t.Errorf("Test 10 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("%02d:%02d-1:00", now.Hour(), now.Minute()+1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 11 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("23:00-%02d:%02d", now.Hour(), now.Minute()-1)
	b = IsSend(s)
	if b {
		t.Errorf("Test 12 false: %s", s.WorkingTime)
	}
	s.WorkingTime = fmt.Sprintf("23:00-%02d:%02d", now.Hour(), now.Minute()+1)
	b = IsSend(s)
	if !b {
		t.Errorf("Test 13 false: %s", s.WorkingTime)
	}

}

func TestsendEmail(t *testing.T) {
	e := emailSettings{}
	e.SMTPServer = "srvmail-exim1.ath.ru:25"
	e.Template = "email.html"
	d := table{}
	sendEmail(e, d)
}
