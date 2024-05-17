package handler

import (
	"currency_viewer/pkg/repository"
	"currency_viewer/pkg/service"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Run() error
}

type handler struct {
	srv    service.Service
	router *chi.Mux
}

func NewHandler(srv service.Service) Handler {
	r := chi.NewRouter()
	h := &handler{
		srv:    srv,
		router: r,
	}
	r.Get("/rate", h.Rate)
	r.Post("/subscribe", h.Subscribe)

	return h
}

func (h *handler) Run() error {
	return http.ListenAndServe(":8080", h.router)
}

func (h *handler) Rate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	rate, err := h.srv.GetUSDRate()
	if err != nil {
		log.Println(err)

		w.WriteHeader(400)
		w.Write([]byte("Invalid status value"))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("%.2f", rate)))
}

func (h *handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	statusCode, message := 200, "OK"
	w.Header().Add("Content-Type", "application/json")

	if err := r.ParseMultipartForm(10<<4); err != nil{
		log.Println(err)
		return
	}
	email := r.FormValue("email")

	err := h.srv.Subscribe(email)
	
	if err == repository.ErrAlreadyExists {
		statusCode, message = 409, err.Error()
	} else if err != nil {
		statusCode, message = 500, "internal server error"
		log.Println(err)
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
