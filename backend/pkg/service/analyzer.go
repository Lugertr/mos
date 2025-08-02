package service

import (
	"center"
	"center/pkg/repository"
)

type AnalyzerService struct {
	repo repository.Analyzer
}

func NewAnalyzerService(repo repository.Analyzer) *AnalyzerService {
	return &AnalyzerService{repo: repo}
}

func (s *AnalyzerService) Create(client center.Analyzer) (int, error) {
	return s.repo.Create(client)
}

func (s *AnalyzerService) GetAll() ([]center.Analyzer, error) {
	return s.repo.GetAll()
}

func (s *AnalyzerService) GetById(clientId int) (center.Analyzer, error) {
	return s.repo.GetById(clientId)
}

func (s *AnalyzerService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *AnalyzerService) Update(clientId int, input center.AnalyzerUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
