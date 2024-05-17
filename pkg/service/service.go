package service

import (
	"currency_viewer/pkg/repository"
)

type Service interface {
	GetUSDRate() (float64, error)
	Subscribe(string) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (srv *service) GetUSDRate() (float64, error) {
	return srv.repo.GetUSDRate()
}

func (srv *service) Subscribe(email string) error{
	return srv.repo.Subscribe(email)
}