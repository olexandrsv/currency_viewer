package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "io"
	_ "net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrAlreadyExists = errors.New("email already exists")
)

type Repository interface {
	GetUSDRate() (float64, error)
	Subscribe(string) error
}

type repo struct {
	db *sql.DB
}

func NewRepository(url string) (Repository, error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}
	return &repo{
		db: db,
	}, nil
}

func (h *repo) GetUSDRate() (float64, error) {
	// resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-UAH")
	// if err != nil{
	// 	return 0, err
	// }

	// data, err := io.ReadAll(resp.Body)
	// if err != nil{
	// 	return 0, err
	// }

	var decoded jsonResponse
	data := `{"USDUAH":{"code":"USD","codein":"UAH","name":"Dólar Americano/Hryvinia Ucraniana","high":"39.3789","low":"39.3452","varBid":"0.0337","pctChange":"0.09","bid":"39.1027","ask":"39.6551","timestamp":"1715943567","create_date":"2024-05-17 07:59:27"}}`

	err := json.Unmarshal([]byte(data), &decoded)
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

func (h *repo) Subscribe(email string) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow("select count(email) from subscribers where email=?", email)

	var emailCount int
	if err = row.Scan(&emailCount); err != nil {
		fmt.Println("here")
		return err
	}

	if emailCount != 0 {
		return ErrAlreadyExists
	}

	_, err = tx.Exec("insert into subscribers (email) values (?)", email)
	if err != nil {
		return err
	}

	return tx.Commit()
}
