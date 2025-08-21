package repository

import (
	"archive"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type CtxUserIDKey struct{}

type Authorization interface {
	CreateUser(ctx context.Context, user archive.User) (int64, error)
	GetUser(ctx context.Context, login, passwordHash string) (archive.User, error)
}

type Authors interface {
	CreateAuthor(ctx context.Context, a archive.Author) (int64, error)
	GetAllAuthors(ctx context.Context) ([]archive.Author, error)
	GetAuthor(ctx context.Context, id int64) (archive.Author, error)
	UpdateAuthor(ctx context.Context, id int64, a archive.Author) error
	DeleteAuthor(ctx context.Context, id int64) error
}

type DocumentTypes interface {
	CreateDocumentType(ctx context.Context, t archive.DocumentType) (int64, error)
	GetAllDocumentTypes(ctx context.Context) ([]archive.DocumentType, error)
	GetDocumentType(ctx context.Context, id int64) (archive.DocumentType, error)
	UpdateDocumentType(ctx context.Context, id int64, t archive.DocumentType) error
	DeleteDocumentType(ctx context.Context, id int64) error
}

type Tags interface {
	CreateTag(ctx context.Context, t archive.Tag) (int64, error)
	GetAllTags(ctx context.Context) ([]archive.Tag, error)
	GetTag(ctx context.Context, id int64) (archive.Tag, error)
	UpdateTag(ctx context.Context, id int64, t archive.Tag) error
	DeleteTag(ctx context.Context, id int64) error
}

type Document interface {
	CreateDocument(ctx context.Context, in archive.DocumentCreateInput) (int64, error)
	SearchDocumentsByTag(ctx context.Context, filter archive.DocumentSearchFilter) ([]archive.DocumentSecure, error)
	GetDocumentByID(ctx context.Context, id int64) (archive.DocumentSecure, error)
	UpdateDocument(ctx context.Context, id int64, in archive.DocumentUpdateInput) error
	DeleteDocument(ctx context.Context, id int64) error

	SetDocumentPermission(ctx context.Context, docID int64, p archive.DocumentPermission) error
	RemoveDocumentPermission(ctx context.Context, docID int64, targetUserID int64) error
}

type Admin interface {
	GetLogsByUser(ctx context.Context, adminID int64, targetUserID int64, start *time.Time, end *time.Time) ([]archive.LogRecord, error)
	GetLogsByTable(ctx context.Context, adminID int64, tableName string, start *time.Time, end *time.Time) ([]archive.LogRecord, error)
	GetLogsByDate(ctx context.Context, adminID int64, start time.Time, end time.Time) ([]archive.LogRecord, error)
}

// Repository aggregates sub-repos
type Repository struct {
	Authorization Authorization
	Authors       Authors
	DocumentTypes DocumentTypes
	Tags          Tags
	Document      Document
	Admin         Admin

	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Authors:       NewAuthorsPostgres(db),
		DocumentTypes: NewDocumentTypesPostgres(db),
		Tags:          NewTagsPostgres(db),
		Document:      NewDocumentPostgres(db),
		Admin:         NewAdminPostgres(db),
		DB:            db,
	}
}
