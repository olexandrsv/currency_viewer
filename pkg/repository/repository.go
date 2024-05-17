package repository

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
}

type repo struct {
	db *sql.DB
}

func NewRepository(url string) (Repository, error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil{
		log.Println(err)
		return nil, errors.New("can't connect to database")
	}
	return &repo{
		db: db,
	}, nil
}