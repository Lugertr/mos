package repository

import (
	"center"

	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user center.UserCreate) (int, error)
	GetUser(username, password string) (center.UserRet, error)
	CheckUser(username string) (bool, error)
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

type Repository struct {
	Authorization
	LabService
	Patient
	ProvidedService
	Analyzer
	InsuranceCompany
	Order
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization:    NewAuthPostgres(db),
		LabService:       NewLabServicePostgres(db),
		Patient:          NewPatientPostgres(db),
		ProvidedService:  NewProvidedServicePostgres(db),
		Analyzer:         NewAnalyzerPostgres(db),
		InsuranceCompany: NewInsuranceCompanyPostgres(db),
		Order:            NewOrderPostgres(db),
	}
}
