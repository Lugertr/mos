package service

import (
	"center"
	"center/pkg/repository"
)

type ProvidedServiceService struct {
	repo repository.ProvidedService
}

func NewProvidedServiceService(repo repository.ProvidedService) *ProvidedServiceService {
	return &ProvidedServiceService{repo: repo}
}

func (s *ProvidedServiceService) Create(client center.ProvidedService) (int, error) {
	return s.repo.Create(client)
}

func (s *ProvidedServiceService) GetAll() ([]center.ProvidedService, error) {
	return s.repo.GetAll()
}

func (s *ProvidedServiceService) GetById(clientId int) (center.ProvidedService, error) {
	return s.repo.GetById(clientId)
}

func (s *ProvidedServiceService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *ProvidedServiceService) Update(clientId int, input center.ProvidedServiceUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
