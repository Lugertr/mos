package service

import (
	"hotel"
	"hotel/pkg/repository"
)

type AppServiceService struct {
	repo repository.AppService
}

func NewAppServiceService(repo repository.AppService) *AppServiceService {
	return &AppServiceService{repo: repo}
}

func (s *AppServiceService) Create(appService hotel.AppService) (int, error) {
	return s.repo.Create(appService)
}

func (s *AppServiceService) GetAll() ([]hotel.AppService, error) {
	return s.repo.GetAll()
}

func (s *AppServiceService) GetById(AppServiceId int) ([]hotel.AppServiceTypeFunc, error) {
	return s.repo.GetById(AppServiceId)
}

func (s *AppServiceService) Delete(AppServiceId int) error {
	return s.repo.Delete(AppServiceId)
}

func (s *AppServiceService) Update(AppServiceId int, input hotel.AppServiceUpdate) error {
	//if err := input.Validate(); err != nil {
	//	return nil
	//}
	return s.repo.Update(AppServiceId, input)
}
