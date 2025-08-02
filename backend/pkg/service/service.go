package service

import (
	"center"
	"center/pkg/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Authorization interface {
	CreateUser(user center.UserCreate) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type LabService interface {
	Create(client center.LabService) (int, error)
	GetAll() ([]center.LabService, error)
	GetById(client_id int) (center.LabService, error)
	Delete(client_id int) error
	Update(client_id int, input center.LabServiceUpdate) error
}

type Patient interface {
	Create(client center.Patient) (int, error)
	GetAll() ([]center.Patient, error)
	GetById(client_id int) (center.Patient, error)
	Delete(client_id int) error
	Update(client_id int, input center.PatientUpdate) error
}

type ProvidedService interface {
	Create(client center.ProvidedService) (int, error)
	GetAll() ([]center.ProvidedService, error)
	GetById(client_id int) (center.ProvidedService, error)
	Delete(client_id int) error
	Update(client_id int, input center.ProvidedServiceUpdate) error
}

type Analyzer interface {
	Create(client center.Analyzer) (int, error)
	GetAll() ([]center.Analyzer, error)
	GetById(client_id int) (center.Analyzer, error)
	Delete(client_id int) error
	Update(client_id int, input center.AnalyzerUpdate) error
}

type InsuranceCompany interface {
	Create(client center.InsuranceCompany) (int, error)
	GetAll() ([]center.InsuranceCompany, error)
	GetById(client_id int) (center.InsuranceCompany, error)
	Delete(client_id int) error
	Update(client_id int, input center.InsuranceCompanyUpdate) error
}

type Order interface {
	Create(client center.Order) (int, error)
	GetAll() ([]center.Order, error)
	GetById(client_id int) (center.Order, error)
	Delete(client_id int) error
	Update(client_id int, input center.OrderUpdate) error
}

type Service struct {
	Authorization
	LabService
	Patient
	ProvidedService
	Analyzer
	InsuranceCompany
	Order
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization:    NewAuthService(repos.Authorization),
		LabService:       NewLabServiceService(repos.LabService),
		Patient:          NewPatientService(repos.Patient),
		ProvidedService:  NewProvidedServiceService(repos.ProvidedService),
		Analyzer:         NewAnalyzerService(repos.Analyzer),
		InsuranceCompany: NewInsuranceCompanyService(repos.InsuranceCompany),
		Order:            NewOrderService(repos.Order),
	}
}
