package service

import (
	"currency_viewer/pkg/repository"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

type Service interface {
	GetUSDRate() (float64, error)
	Subscribe(string) error
}

type service struct {
	repo            repository.Repository
	login, password string
}

type Config struct {
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

func NewService(repo repository.Repository) (Service, func(), error) {
	inidata, err := ini.Load("../config/config.ini")
	if err != nil {
		return nil, func() {}, err
	}
	section := inidata.Section("email")

	done := make(chan struct{})
	srv := &service{
		repo: repo,
	}
	srv.login = section.Key("login").String()
	srv.password = section.Key("password").String()

	go srv.notificationLoop(done)
	return srv, func() {
		done <- struct{}{}
		close(done)
	}, nil
}

func (srv *service) GetUSDRate() (float64, error) {
	return srv.repo.GetUSDRate()
}

func (srv *service) Subscribe(email string) error {
	return srv.repo.Subscribe(email)
}

func (srv *service) notificationLoop(done <-chan struct{}) {
	isTime, err := srv.isNotificationTime()
	if err != nil {
		log.Println(err)
		return
	}

	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	timeToSend := time.After(time.Until(tomorrow))

	if isTime {
		if err := srv.sendEmails(); err != nil {
			log.Println(err)
			return
		}
		timeToSend = time.After(time.Hour * 24)
	}

	for {
		select {
		case <-timeToSend:
			if err := srv.sendEmails(); err != nil {
				log.Println(err)
				return
			}
			timeToSend = time.After(time.Hour * 24)
		case <-done:
			return
		}
	}
}

func (srv *service) isNotificationTime() (bool, error) {
	date, err := srv.repo.GetLastNotificationDate()
	if err != nil {
		return false, err
	}

	if date == "" {
		return true, nil
	}

	template := "1/2/2006"
	notificationDate, err := time.Parse(template, date)
	if err != nil {
		return false, err
	}
	now := time.Now()
	if notificationDate.Day() != now.Day() || notificationDate.Month() != now.Month() ||
		notificationDate.Year() != now.Year() {
		return true, nil
	}

	return false, nil
}

func (srv *service) sendEmails() error {
	emails, err := srv.repo.GetAllEmails()
	if err != nil {
		return err
	}

	for _, email := range emails {
		if err := srv.sendEmail(email); err != nil {
			return err
		}
	}

	now := time.Now()
	date := strconv.Itoa(int(now.Month())) + "/" + strconv.Itoa(now.Day()) + "/" + strconv.Itoa(now.Year())
	srv.repo.UpdateLastNotificationDate(date)

	return nil
}

func (srv *service) sendEmail(email string) error {
	rate, err := srv.GetUSDRate()
	if err != nil {
		return err
	}
	auth := smtp.PlainAuth("", srv.login, srv.password, "smtp.gmail.com")

	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"USD-UAH Rate\r\n" +
		"\r\n" +
		fmt.Sprintf("%.2f", rate) + "\r\n")

	err = smtp.SendMail("smtp.gmail.com:587", auth, srv.login, to, msg)
	if err != nil {
		return err
	}

	return nil
}
