package service

import (
	"center"
	"center/pkg/repository"
)

type PatientService struct {
	repo repository.Patient
}

func NewPatientService(repo repository.Patient) *PatientService {
	return &PatientService{repo: repo}
}

func (s *PatientService) Create(client center.Patient) (int, error) {
	return s.repo.Create(client)
}

func (s *PatientService) GetAll() ([]center.Patient, error) {
	return s.repo.GetAll()
}

func (s *PatientService) GetById(clientId int) (center.Patient, error) {
	return s.repo.GetById(clientId)
}

func (s *PatientService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *PatientService) Update(clientId int, input center.PatientUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
