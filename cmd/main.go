package main

import (
	"currency_viewer/pkg/handler"
	"currency_viewer/pkg/repository"
	"currency_viewer/pkg/service"
	"log"
)

func main(){
	repo, err := repository.NewRepository("../../subscribers.db")
	if err != nil{
		log.Println(err)
		return
	}
	srv := service.NewService(repo)
	h := handler.NewHandler(srv)

	if err = h.Run(); err != nil{
		log.Println(err)
		return
	}
}