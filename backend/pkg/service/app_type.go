package service

import (
	"hotel"
	"hotel/pkg/repository"
)

type AppTypeService struct {
	repo repository.AppType
}

func NewAppTypeService(repo repository.AppType) *AppTypeService {
	return &AppTypeService{repo: repo}
}

func (s *AppTypeService) Create(appType hotel.AppType) (int, error) {
	return s.repo.Create(appType)
}

func (s *AppTypeService) GetAll() ([]hotel.AppType, error) {
	return s.repo.GetAll()
}

func (s *AppTypeService) GetById(appTypeId int) (hotel.AppType, error) {
	return s.repo.GetById(appTypeId)
}

func (s *AppTypeService) Delete(appTypeId int) error {
	return s.repo.Delete(appTypeId)
}

func (s *AppTypeService) Update(appTypeId int, input hotel.AppTypeUpdate) error {
	//if err := input.Validate(); err != nil {
	//	return nil
	//}
	return s.repo.Update(appTypeId, input)
}
