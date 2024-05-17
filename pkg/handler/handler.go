package handler

import (
	"currency_viewer/pkg/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Run() error
}

type handler struct {
	srv service.Service
	router *chi.Mux
}

func NewHandler(srv service.Service) Handler{
	r := chi.NewRouter()
	return &handler{
		srv: srv,
		router: r,
	}
}

func (h *handler) Run() error {
	return http.ListenAndServe(":8080", h.router)
}

