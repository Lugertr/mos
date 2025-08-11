package repository

import (
	"hotel"

	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user hotel.User) (int, error)
	GetUser(username, password string) (hotel.User, error)
}

type Client interface {
	Create(client hotel.Client) (int, error)
	GetAll() ([]hotel.Client, error)
	GetById(client_id int) (hotel.ClientFunc, error)
	Delete(client_id int) error
	Update(client_id int, input hotel.ClientUpdate) error
}

type App interface {
	Create(app hotel.App) (int, error)
	GetAll() ([]hotel.App, error)
	GetById(app_id int) ([]hotel.App, error)
	Delete(app_id int) error
	Update(app_id int, input hotel.AppUpdate) error
}

type AppType interface {
	Create(appType hotel.AppType) (int, error)
	GetAll() ([]hotel.AppType, error)
	GetById(appTypeId int) (hotel.AppType, error)
	Delete(appTypeId int) error
	Update(appTypeId int, input hotel.AppTypeUpdate) error
}

type AppService interface {
	Create(appService hotel.AppService) (int, error)
	GetAll() ([]hotel.AppService, error)
	GetById(AppServiceId int) ([]hotel.AppServiceTypeFunc, error)
	Delete(AppServiceId int) error
	Update(AppServiceId int, input hotel.AppServiceUpdate) error
}

type AppServiceType interface {
	Create(appServiceType hotel.AppServiceType) (int, error)
	GetAll() ([]hotel.AppServiceType, error)
	GetById(appServiceTypeId int) (hotel.AppServiceType, error)
	Delete(appServiceTypeId int) error
	Update(appServiceTypeId int, input hotel.AppServiceTypeUpdate) error
}

type Repository struct {
	Authorization
	Client
	App
	AppType
	AppService
	AppServiceType
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization:  NewAuthPostgres(db),
		Client:         NewClientPostgres(db),
		App:            NewAppPostgres(db),
		AppType:        NewAppTypePostgres(db),
		AppService:     NewAppServicePostgres(db),
		AppServiceType: NewAppServiceTypePostgres(db),
	}
}
