package service

import (
	"hotel"
	"hotel/pkg/repository"
)

type AppServices struct {
	repo repository.App
}

func NewAppService(repo repository.App) *AppServices {
	return &AppServices{repo: repo}
}

func (s *AppServices) Create(app hotel.App) (int, error) {
	return s.repo.Create(app)
}

func (s *AppServices) GetAll() ([]hotel.App, error) {
	return s.repo.GetAll()
}

func (s *AppServices) GetById(appId int) ([]hotel.App, error) {
	return s.repo.GetById(appId)
}

func (s *AppServices) Delete(appId int) error {
	return s.repo.Delete(appId)
}

func (s *AppServices) Update(appId int, input hotel.AppUpdate) error {
	//if err := input.Validate(); err != nil {
	//	return nil
	//}
	return s.repo.Update(appId, input)
}
