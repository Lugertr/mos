package service

import (
	"center"
	"center/pkg/repository"
)

type LabServiceService struct {
	repo repository.LabService
}

func NewLabServiceService(repo repository.LabService) *LabServiceService {
	return &LabServiceService{repo: repo}
}

func (s *LabServiceService) Create(client center.LabService) (int, error) {
	return s.repo.Create(client)
}

func (s *LabServiceService) GetAll() ([]center.LabService, error) {
	return s.repo.GetAll()
}

func (s *LabServiceService) GetById(clientId int) (center.LabService, error) {
	return s.repo.GetById(clientId)
}

func (s *LabServiceService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *LabServiceService) Update(clientId int, input center.LabServiceUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
