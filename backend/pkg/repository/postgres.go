package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	// registration / auth
	fnRegisterUser  = "fn_register_user"
	fnAuthorizeUser = "fn_authorize_user"

	// document CRUD / permissions
	fnAddDocument              = "fn_add_document"
	fnUpdateDocument           = "fn_update_document"
	fnDeleteDocument           = "fn_delete_document"
	fnSetDocumentPermission    = "fn_set_document_permission"
	fnRemoveDocumentPermission = "fn_remove_document_permission"
	fnGetDocumentsForUser      = "fn_get_documents_for_user"
	fnGetDocumentByID          = "fn_get_document_by_id"

	// logs
	fnGetLogsByUser  = "fn_get_logs_by_user"
	fnGetLogsByTable = "fn_get_logs_by_table"
	fnGetLogsByDate  = "fn_get_logs_by_date"

	// internals
	internalGetOrCreateAuthor = "_internal_get_or_create_author"
	internalGetOrCreateTag    = "_internal_get_or_create_tag"

	documentTypesTable = "document_types"
	authorsTable       = "authors"
	tagsTable          = "tags"
	documentTagsTable  = "document_tags"
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
