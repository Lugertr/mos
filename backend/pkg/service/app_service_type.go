package service

import (
	"hotel"
	"hotel/pkg/repository"
)

type AppServiceTypeService struct {
	repo repository.AppServiceType
}

func NewAppServiceTypeService(repo repository.AppServiceType) *AppServiceTypeService {
	return &AppServiceTypeService{repo: repo}
}

func (s *AppServiceTypeService) Create(appServiceType hotel.AppServiceType) (int, error) {
	return s.repo.Create(appServiceType)
}

func (s *AppServiceTypeService) GetAll() ([]hotel.AppServiceType, error) {
	return s.repo.GetAll()
}

func (s *AppServiceTypeService) GetById(appServiceTypeId int) (hotel.AppServiceType, error) {
	return s.repo.GetById(appServiceTypeId)
}

func (s *AppServiceTypeService) Delete(appServiceTypeId int) error {
	return s.repo.Delete(appServiceTypeId)
}

func (s *AppServiceTypeService) Update(appServiceTypeId int, input hotel.AppServiceTypeUpdate) error {
	//if err := input.Validate(); err != nil {
	//	return nil
	//}
	return s.repo.Update(appServiceTypeId, input)
}
