package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	_ "io"
	"net/http"
	_ "net/http"
	"os"
	"strconv"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrAlreadyExists = errors.New("email already exists")
)

type Repository interface {
	GetUSDRate() (float64, error)
	Subscribe(string) error
	GetAllEmails() ([]string, error)
	GetLastNotificationDate() (string, error)
	UpdateLastNotificationDate(string) error
}

type repo struct {
	db *sql.DB
}

func NewRepository(url string) (Repository, func(), error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, func() {}, err
	}
	close := func() {
		db.Close()
	}

	fs := os.DirFS("./../")
	sourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		panic(err)
	}

	databaseDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, close, err
	}

	m, err := migrate.NewWithInstance(
		"file", sourceDriver,
		"currency_viewer", databaseDriver)
	if err != nil {
		return nil, close, err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, close, err
	}
	return &repo{
		db: db,
	}, close, nil
}

func (r *repo) GetUSDRate() (float64, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-UAH")
	if err != nil{
		return 0, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil{
		return 0, err
	}

	var decoded jsonResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return 0, err
	}

	rate, err := strconv.ParseFloat(decoded.Header.Rate, 64)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

type jsonResponse struct {
	Header data `json:"USDUAH"`
}

type data struct {
	Rate string `json:"high"`
}

func (r *repo) Subscribe(email string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow("select count(email) from subscribers where email=?", email)

	var emailCount int
	if err = row.Scan(&emailCount); err != nil {
		tx.Rollback()
		return err
	}

	if emailCount != 0 {
		tx.Rollback()
		return ErrAlreadyExists
	}

	_, err = tx.Exec("insert into subscribers (email) values (?)", email)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repo) GetAllEmails() ([]string, error) {
	var emails []string
	rows, err := r.db.Query("select email from subscribers")
	if err != nil {
		return nil, err
	}

	var email string
	for rows.Next() {
		if err = rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (r *repo) GetLastNotificationDate() (string, error) {
	row := r.db.QueryRow("select time from notifications limit 1")

	var date string
	err := row.Scan(&date)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return date, nil
}

func (r *repo) UpdateLastNotificationDate(date string) error {
	_, err := r.db.Exec("delete from notifications; insert into notifications (time) values (?)", date)
	return err
}
