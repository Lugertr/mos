package service

import (
	"center"
	"center/pkg/repository"
)

type InsuranceCompanyService struct {
	repo repository.InsuranceCompany
}

func NewInsuranceCompanyService(repo repository.InsuranceCompany) *InsuranceCompanyService {
	return &InsuranceCompanyService{repo: repo}
}

func (s *InsuranceCompanyService) Create(client center.InsuranceCompany) (int, error) {
	return s.repo.Create(client)
}

func (s *InsuranceCompanyService) GetAll() ([]center.InsuranceCompany, error) {
	return s.repo.GetAll()
}

func (s *InsuranceCompanyService) GetById(clientId int) (center.InsuranceCompany, error) {
	return s.repo.GetById(clientId)
}

func (s *InsuranceCompanyService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *InsuranceCompanyService) Update(clientId int, input center.InsuranceCompanyUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
