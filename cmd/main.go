package main

import (
	"currency_viewer/pkg/handler"
	"currency_viewer/pkg/repository"
	"currency_viewer/pkg/service"
	"log"
)

func main() {
	repo, close, err := repository.NewRepository("../currency_viewer.db")
	defer close()
	if err != nil {
		log.Println(err)
		return
	}
	srv, cancel, err := service.NewService(repo)
	defer cancel()
	if err != nil {
		log.Println(err)
		return
	}
	h := handler.NewHandler(srv)

	if err = h.Run(); err != nil {
		log.Println(err)
		return
	}
}
