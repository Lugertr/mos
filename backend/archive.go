package archive

import (
	"encoding/json"
	"time"
)

// --- Справочники ---------------------------------------------------------

type DocumentType struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Tag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
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

type FileMeta struct {
	Provider     string `json:"provider,omitempty"`      // "s3"|"minio"|"local"
	Bucket       string `json:"bucket,omitempty"`        // bucket name
	Key          string `json:"key,omitempty"`           // object key
	Mime         string `json:"mime,omitempty"`          // content-type
	Size         int64  `json:"size,omitempty"`          // bytes
	Sha256       string `json:"sha256,omitempty"`        // hex checksum
	StorageClass string `json:"storage_class,omitempty"` // optional
	ThumbKey     string `json:"thumbnail_key,omitempty"` // optional
}

// Document — основная сущность документов.
// Поля указателями там, где в схеме допускается NULL.
type Document struct {
	ID           int64            `db:"id" json:"id"`
	Title        string           `db:"title" json:"title"`
	Privacy      PrivacyType      `db:"privacy" json:"privacy"`
	CreatedAt    time.Time        `db:"created_at" json:"created_at"`
	CreatedBy    *int64           `db:"created_by" json:"created_by,omitempty"`
	UpdatedAt    *time.Time       `db:"updated_at" json:"updated_at,omitempty"`
	UpdatedBy    *int64           `db:"updated_by" json:"updated_by,omitempty"`
	DocumentDate *time.Time       `db:"document_date" json:"document_date,omitempty"`
	Author       *string          `db:"author" json:"author,omitempty"`
	TypeID       *int64           `db:"type_id" json:"type_id,omitempty"`
	FileMeta     *FileMeta        `db:"file_meta" json:"file_meta,omitempty"`
	GeoJSON      *json.RawMessage `db:"geojson" json:"geojson,omitempty"`
	Geom         *string          `db:"geom" json:"geom,omitempty"`
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
	Author   string `json:"author"`    // автор (имя)
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
	Changes    json.RawMessage `db:"changes" json:"changes,omitempty"`
}

// --- Дополнительные , удобные для сервисов/handler'ов -------------------

// DocumentSecure — результат security-функций (fn_get_document_by_id / fn_get_documents_for_user)
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
	Author            *string     `db:"author" json:"author,omitempty"`
	TypeID            *int64      `db:"type_id" json:"type_id,omitempty"`
	TypeName          *string     `db:"type_name" json:"type_name,omitempty"`
	Tags              []string    `db:"tags" json:"tags,omitempty"`
	Viewers           []int64     `db:"viewers" json:"viewers,omitempty"`
	Editors           []int64     `db:"editors" json:"editors,omitempty"`
	CanRequesterEdit  bool        `db:"can_requester_edit" json:"can_requester_edit"`
	Geom              *string     `db:"geom" json:"geom,omitempty"` // ST_AsGeoJSON(geom)
	FileMeta          *FileMeta   `db:"file_meta" json:"file_meta,omitempty"`
	DownloadURL       string      `json:"download_url,omitempty"`
}

// DocumentCreateInput — удобная структура для передачи данных из handler->service
type DocumentCreateInput struct {
	Title        string
	Privacy      PrivacyType
	DocumentDate *time.Time
	Author       *string
	TypeID       *int64
	FileMeta     *FileMeta
	GeoJSON      *json.RawMessage
	Tags         []string
	CreatorID    int64
}

// DocumentUpdateInput — для обновления документа
type DocumentUpdateInput struct {
	DocumentID   int64
	Title        *string
	Privacy      *PrivacyType
	DocumentDate *time.Time
	Author       *string
	TypeID       *int64
	FileMeta     *FileMeta
	GeoJSON      *json.RawMessage
	Tags         *[]string
	UpdaterID    int64
}
