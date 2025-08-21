package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	usersTable         = "users"
	authorsTable       = "authors"
	documentTypesTable = "document_types"
	tagsTable          = "tags"
	documentsTable     = "documents"
	documentTagsTable  = "document_tags"
	documentPermsTable = "document_permissions"
	logsTable          = "logs"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
	Timeout  time.Duration
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}
	if cfg.Timeout > 0 {
		db.SetConnMaxLifetime(cfg.Timeout)
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
