package service

import (
	"archive"
	"context"
	"time"
)

// Authorization сервис (аутентификация)
type Authorization interface {
	CreateUser(ctx context.Context, user archive.User) (int64, error)
	GenerateToken(ctx context.Context, username, password string) (string, error)
	RefreshToken(ctx context.Context, accessToken string) (string, error)
	ParseToken(ctx context.Context, token string) (int64, error)
}

// Authors сервис (справочник authors)
type Authors interface {
	CreateAuthor(ctx context.Context, in archive.AuthorCreate) (int64, error)
	GetAllAuthors(ctx context.Context) ([]archive.Author, error)
	GetAuthor(ctx context.Context, id int64) (archive.Author, error)
	UpdateAuthor(ctx context.Context, id int64, in archive.Author) error
	DeleteAuthor(ctx context.Context, id int64) error
}

// DocumentTypes сервис (справочник document_types)
type DocumentTypes interface {
	CreateDocumentType(ctx context.Context, in archive.DocumentTypeCreate) (int64, error)
	GetAllDocumentTypes(ctx context.Context) ([]archive.DocumentType, error)
	GetDocumentType(ctx context.Context, id int64) (archive.DocumentType, error)
	UpdateDocumentType(ctx context.Context, id int64, in archive.DocumentType) error
	DeleteDocumentType(ctx context.Context, id int64) error
}

// Tags сервис (справочник tags)
type Tags interface {
	CreateTag(ctx context.Context, in archive.TagCreate) (int64, error)
	GetAllTags(ctx context.Context) ([]archive.Tag, error)
	GetTag(ctx context.Context, id int64) (archive.Tag, error)
	UpdateTag(ctx context.Context, id int64, in archive.Tag) error
	DeleteTag(ctx context.Context, id int64) error
}

// Document и Admin оставляем как прежде (с контекстом)
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
