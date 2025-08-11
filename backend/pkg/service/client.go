package service

import (
	"hotel"
	"hotel/pkg/repository"
)

type ClientService struct {
	repo repository.Client
}

func NewClientService(repo repository.Client) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) Create(client hotel.Client) (int, error) {
	return s.repo.Create(client)
}

func (s *ClientService) GetAll() ([]hotel.Client, error) {
	return s.repo.GetAll()
}

func (s *ClientService) GetById(clientId int) (hotel.ClientFunc, error) {
	return s.repo.GetById(clientId)
}

func (s *ClientService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *ClientService) Update(clientId int, input hotel.ClientUpdate) error {
	//if err := input.Validate(); err != nil {
	//	return nil
	//}
	return s.repo.Update(clientId, input)
}
