package archive

import (
	"encoding/json"
	"time"
)

// --- Справочники ---------------------------------------------------------

type Author struct {
	ID       int64  `db:"id" json:"id"`
	FullName string `db:"full_name" json:"full_name"`
}

type DocumentType struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type AuthorCreate struct {
	FullName string `json:"full_name" validate:"required,min=2"`
}

type TagCreate struct {
	Name string `json:"name" validate:"required,min=1"`
}

type DocumentTypeCreate struct {
	Name string `json:"name" validate:"required,min=1"`
}

// --- Таблица documents ---------------------------------------------------

// Privacy хранится в БД как enum privacy_type ('public'|'private')
type PrivacyType string

const (
	PrivacyPublic  PrivacyType = "public"
	PrivacyPrivate PrivacyType = "private"
)

// Document — основная сущность документов.
// Поля делаются указателями там, где в схеме допускается NULL.
type Document struct {
	ID           int64       `db:"id" json:"id"`
	Title        string      `db:"title" json:"title"`
	Privacy      PrivacyType `db:"privacy" json:"privacy"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	CreatedBy    *int64      `db:"created_by" json:"created_by,omitempty"`
	UpdatedAt    *time.Time  `db:"updated_at" json:"updated_at,omitempty"`
	UpdatedBy    *int64      `db:"updated_by" json:"updated_by,omitempty"`
	DocumentDate *time.Time  `db:"document_date" json:"document_date,omitempty"`
	AuthorID     *int64      `db:"author_id" json:"author_id,omitempty"`
	TypeID       *int64      `db:"type_id" json:"type_id,omitempty"`
	// файл в bytea
	File []byte `db:"file_bytea" json:"file,omitempty"`
	// geojson хранится в БД как jsonb
	GeoJSON json.RawMessage `db:"geojson" json:"geojson,omitempty"`
	// geom — geometry(Geometry,4326). При SELECT удобно получать в виде GeoJSON/ WKT;
	// тут — строка (например, результат ST_AsGeoJSON(geom) или ST_AsText(geom))
	Geom *string `db:"geom" json:"geom,omitempty"`
}

// --- Связующие таблицы ---------------------------------------------------

type DocumentTag struct {
	DocumentID int64 `db:"document_id" json:"document_id"`
	TagID      int64 `db:"tag_id" json:"tag_id"`
}

type DocumentPermission struct {
	DocumentID int64 `db:"document_id" json:"document_id"`
	UserID     int64 `db:"user_id" json:"user_id"`
	CanView    bool  `db:"can_view" json:"can_view"`
	CanEdit    bool  `db:"can_edit" json:"can_edit"`
}

// DocumentSearchFilter — фильтр для поиска документов (используется в handlers/services)
type DocumentSearchFilter struct {
	Tag      string `json:"tag"`       // тег
	Author   string `json:"author"`    // автор (имя) или author_id в виде строки
	Type     string `json:"type"`      // тип документа (имя) или type_id
	DateFrom string `json:"date_from"` // диапазон дат — левые/правые границы (строки парсятся в сервисе/handler)
	DateTo   string `json:"date_to"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// --- Логи ----------------------------------------------------------------

type LogRecord struct {
	ID         int64           `db:"id" json:"id"`
	Action     string          `db:"action" json:"action"`
	TableName  string          `db:"table_name" json:"table_name"`
	RecordID   *int64          `db:"record_id" json:"record_id,omitempty"`
	UserID     *int64          `db:"user_id" json:"user_id,omitempty"`
	UserLogin  *string         `db:"user_login" json:"user_login,omitempty"`
	ActionTime time.Time       `db:"action_time" json:"action_time"`
	Changes    json.RawMessage `db:"changes" json:"changes,omitempty"` // jsonb
}

// --- Дополнительные , удобные для сервисов/handler'ов -------------------

// DocumentSecure — результат fn_get_document_secure / fn_get_documents_by_tag_secure
// Поля соответствуют возвращаемым колонкам функции.
type DocumentSecure struct {
	DocID             int64       `db:"doc_id" json:"doc_id"`
	Title             string      `db:"title" json:"title"`
	Privacy           PrivacyType `db:"privacy" json:"privacy"`
	CreatedAt         time.Time   `db:"created_at" json:"created_at"`
	CreatedBy         *int64      `db:"created_by" json:"created_by,omitempty"`
	CreatedByLogin    *string     `db:"created_by_login" json:"created_by_login,omitempty"`
	CreatedByFullName *string     `db:"created_by_full_name" json:"created_by_full_name,omitempty"`
	UpdatedAt         *time.Time  `db:"updated_at" json:"updated_at,omitempty"`
	UpdatedBy         *int64      `db:"updated_by" json:"updated_by,omitempty"`
	UpdatedByLogin    *string     `db:"updated_by_login" json:"updated_by_login,omitempty"`
	UpdatedByFullName *string     `db:"updated_by_full_name" json:"updated_by_full_name,omitempty"`
	DocumentDate      *time.Time  `db:"document_date" json:"document_date,omitempty"`
	AuthorID          *int64      `db:"author_id" json:"author_id,omitempty"`
	AuthorName        *string     `db:"author_name" json:"author_name,omitempty"`
	TypeID            *int64      `db:"type_id" json:"type_id,omitempty"`
	TypeName          *string     `db:"type_name" json:"type_name,omitempty"`
	Tags              []string    `db:"tags" json:"tags,omitempty"`
	Viewers           []int64     `db:"viewers" json:"viewers,omitempty"`
	Editors           []int64     `db:"editors" json:"editors,omitempty"`
	CanRequesterEdit  bool        `db:"can_requester_edit" json:"can_requester_edit"`
	Geom              *string     `db:"geom" json:"geom,omitempty"`
}

// DocumentCreateInput — удобная структура для передачи данных из handler->service
type DocumentCreateInput struct {
	Title        string
	Privacy      PrivacyType
	DocumentDate *time.Time
	AuthorID     *int64
	AuthorName   *string
	TypeID       *int64
	File         []byte
	GeoJSON      json.RawMessage
	Tags         []string
	CreatorID    int64
}

// DocumentUpdateInput — для обновления документа
type DocumentUpdateInput struct {
	DocumentID   int64
	Title        *string
	Privacy      *PrivacyType
	DocumentDate *time.Time
	AuthorID     *int64
	AuthorName   *string
	TypeID       *int64
	File         *[]byte
	GeoJSON      *json.RawMessage
	Tags         *[]string
	UpdaterID    int64
}
