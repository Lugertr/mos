package service

import (
	"center"
	"center/pkg/repository"
)

type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Create(client center.Order) (int, error) {
	return s.repo.Create(client)
}

func (s *OrderService) GetAll() ([]center.Order, error) {
	return s.repo.GetAll()
}

func (s *OrderService) GetById(clientId int) (center.Order, error) {
	return s.repo.GetById(clientId)
}

func (s *OrderService) Delete(clientId int) error {
	return s.repo.Delete(clientId)
}

func (s *OrderService) Update(clientId int, input center.OrderUpdate) error {
	if err := input.Validate(); err != nil {
		return nil
	}
	return s.repo.Update(clientId, input)
}
