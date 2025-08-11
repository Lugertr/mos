package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	usersTable          = "users"
	clientTable         = "client_table"
	appTable            = "app_table"
	appTypeTable        = "app_type_table"
	appServiceTable     = "service_table"
	appServiceTypeTable = "service_type_table"
	hotelListsTable     = "hotel_lists"
	usersListsTable     = "users_list"
	hotelItemsTable     = "hotel_items"
	listsItemsTable     = "lists_items"
	appFunc             = "get_today_tomorrow()"
	serviceFunc         = "get_service_types()"
	clientFunc          = "get_client_info($1)"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
