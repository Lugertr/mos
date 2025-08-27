package repository

import (
	"archive"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type DocumentPostgres struct {
	db *sqlx.DB
}

func NewDocumentPostgres(db *sqlx.DB) *DocumentPostgres {
	return &DocumentPostgres{db: db}
}

func userIDFromCtx(ctx context.Context) (int64, bool) {
	v := ctx.Value(CtxUserIDKey{})
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case *int64:
		if t == nil {
			return 0, false
		}
		return *t, true
	default:
		return 0, false
	}
}

// CreateDocument -> fn_add_document
func (r *DocumentPostgres) CreateDocument(ctx context.Context, in archive.DocumentCreateInput) (int64, error) {
	var id int64

	var geojson interface{}
	if len(in.GeoJSON) > 0 {
		if !json.Valid(in.GeoJSON) {
			return 0, fmt.Errorf("invalid geojson")
		}
		geojson = in.GeoJSON
	} else {
		geojson = nil
	}

	// SELECT fn_add_document($1, $2, ... )
	query := `SELECT ` + fnAddDocument + `($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	err := r.db.QueryRowxContext(ctx, query,
		in.CreatorID,
		in.Title,
		in.DocumentDate,
		in.AuthorID,
		in.AuthorName,
		in.TypeID,
		in.File,
		geojson,
		[]string(in.Tags),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SearchDocumentsByTag -> в новой БД нет fn_get_documents_by_tag_secure.
// Используем fn_get_documents_for_user(p_requester_id) — возвращает список документов доступных пользователю.
// Параметр filter.Tag/Author/Type не учитывается (вы можете расширить БД/функцию при необходимости).
func (r *DocumentPostgres) SearchDocumentsByTag(ctx context.Context, filter archive.DocumentSearchFilter) ([]archive.DocumentSecure, error) {
	// Запрос к security-функции, которая возвращает таблицу для пользователя.
	query := `
SELECT
  id,
  title,
  privacy,
  updated_at,
  document_date,
  author_id,
  type_id,
  geojson,
  can_edit,
  is_author
FROM ` + fnGetDocumentsForUser + `($1)
ORDER BY COALESCE(updated_at, now()) DESC
`

	var requester interface{} = nil
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}

	// промежуточная структура для сканирования результата fn_get_documents_for_user
	type listRow struct {
		ID           int64               `db:"id"`
		Title        string              `db:"title"`
		Privacy      archive.PrivacyType `db:"privacy"`
		UpdatedAt    sql.NullTime        `db:"updated_at"`
		DocumentDate *time.Time          `db:"document_date"`
		AuthorID     *int64              `db:"author_id"`
		TypeID       *int64              `db:"type_id"`
		GeoJSON      json.RawMessage     `db:"geojson"`
		CanEdit      bool                `db:"can_edit"`
		IsAuthor     bool                `db:"is_author"`
	}

	var rows []listRow
	if err := r.db.SelectContext(ctx, &rows, query, requester); err != nil {
		return nil, err
	}

	out := make([]archive.DocumentSecure, 0, len(rows))
	for _, r0 := range rows {
		var updatedAtPtr *time.Time
		if r0.UpdatedAt.Valid {
			t := r0.UpdatedAt.Time
			updatedAtPtr = &t
		}
		// Map to DocumentSecure — заполним те поля, которые есть в ответе функции.
		ds := archive.DocumentSecure{
			DocID:            r0.ID,
			Title:            r0.Title,
			Privacy:          r0.Privacy,
			UpdatedAt:        updatedAtPtr,
			DocumentDate:     r0.DocumentDate,
			AuthorID:         r0.AuthorID,
			TypeID:           r0.TypeID,
			Tags:             nil, // fn_get_documents_for_user не возвращает tags
			Viewers:          nil,
			Editors:          nil,
			CanRequesterEdit: r0.CanEdit,
		}
		// If geojson exists, try to put it into Geom as GeoJSON string (optional)
		if len(r0.GeoJSON) > 0 {
			s := string(r0.GeoJSON)
			ds.Geom = &s
		}
		out = append(out, ds)
	}

	// Apply offset/limit at app level if requested (fn doesn't support it here)
	start := filter.Offset
	if start < 0 {
		start = 0
	}
	end := len(out)
	if filter.Limit > 0 {
		if start+filter.Limit < end {
			end = start + filter.Limit
		}
	}
	if start > len(out) {
		return []archive.DocumentSecure{}, nil
	}
	return out[start:end], nil
}

func nullString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// GetDocumentByID -> fn_get_document_by_id(document_id, requester_id)
func (r *DocumentPostgres) GetDocumentByID(ctx context.Context, id int64) (archive.DocumentSecure, error) {
	var out archive.DocumentSecure
	// fn_get_document_by_id returns a different set of columns (see SQL). We'll query and map available columns.
	query := `
SELECT
  id,
  title,
  privacy,
  created_at,
  created_by,
  updated_at,
  updated_by,
  document_date,
  author_id,
  type_id,
  file_bytea,
  geojson,
  ST_AsGeoJSON(geom) as geom,
  can_edit
FROM ` + fnGetDocumentByID + `($1,$2)
LIMIT 1
`
	var requester interface{} = nil
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}
	// промежуточ структура для сканирования
	type docRow struct {
		ID           int64               `db:"id"`
		Title        string              `db:"title"`
		Privacy      archive.PrivacyType `db:"privacy"`
		CreatedAt    time.Time           `db:"created_at"`
		CreatedBy    *int64              `db:"created_by"`
		UpdatedAt    sql.NullTime        `db:"updated_at"`
		UpdatedBy    *int64              `db:"updated_by"`
		DocumentDate *time.Time          `db:"document_date"`
		AuthorID     *int64              `db:"author_id"`
		TypeID       *int64              `db:"type_id"`
		File         []byte              `db:"file_bytea"`
		GeoJSON      json.RawMessage     `db:"geojson"`
		Geom         *string             `db:"geom"`
		CanEdit      bool                `db:"can_edit"`
	}

	var row docRow
	if err := r.db.GetContext(ctx, &row, query, id, requester); err != nil {
		if err == sql.ErrNoRows {
			return archive.DocumentSecure{}, nil
		}
		return archive.DocumentSecure{}, err
	}

	var updatedAtPtr *time.Time
	if row.UpdatedAt.Valid {
		t := row.UpdatedAt.Time
		updatedAtPtr = &t
	}

	out = archive.DocumentSecure{
		DocID:            row.ID,
		Title:            row.Title,
		Privacy:          row.Privacy,
		CreatedAt:        row.CreatedAt,
		CreatedBy:        row.CreatedBy,
		UpdatedAt:        updatedAtPtr,
		UpdatedBy:        row.UpdatedBy,
		DocumentDate:     row.DocumentDate,
		AuthorID:         row.AuthorID,
		TypeID:           row.TypeID,
		Tags:             nil, // not returned by fn_get_document_by_id
		Viewers:          nil,
		Editors:          nil,
		CanRequesterEdit: row.CanEdit,
		Geom:             row.Geom,
	}

	// file/geojson are available in row.File / row.GeoJSON if needed by caller (but DocumentSecure doesn't expose file)
	_ = row.File
	_ = row.GeoJSON

	return out, nil
}

// UpdateDocument -> fn_update_document
func (r *DocumentPostgres) UpdateDocument(ctx context.Context, id int64, in archive.DocumentUpdateInput) error {
	if in.DocumentID == 0 {
		in.DocumentID = id
	}
	if in.UpdaterID == 0 {
		if uid, ok := userIDFromCtx(ctx); ok {
			in.UpdaterID = uid
		} else {
			return fmt.Errorf("updater id required")
		}
	}

	var titleParam interface{}
	if in.Title != nil {
		titleParam = *in.Title
	} else {
		titleParam = nil
	}
	var fileParam interface{}
	if in.File != nil {
		fileParam = *in.File
	} else {
		fileParam = nil
	}
	var geojsonParam interface{}
	if in.GeoJSON != nil {
		// validate JSON if present
		if !json.Valid(*in.GeoJSON) {
			return fmt.Errorf("invalid geojson")
		}
		geojsonParam = *in.GeoJSON
	} else {
		geojsonParam = nil
	}
	var tagsParam interface{}
	if in.Tags != nil {
		tagsParam = *in.Tags
	} else {
		tagsParam = nil
	}
	var privacyParam interface{}
	if in.Privacy != nil {
		privacyParam = string(*in.Privacy)
	} else {
		privacyParam = nil
	}

	query := `SELECT ` + fnUpdateDocument + `($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, query,
		in.DocumentID,
		in.UpdaterID,
		titleParam,
		in.DocumentDate,
		in.AuthorID,
		in.AuthorName,
		in.TypeID,
		fileParam,
		geojsonParam,
		tagsParam,
		privacyParam,
	)
	return err
}

// DeleteDocument -> fn_delete_document(document_id, user_id)
func (r *DocumentPostgres) DeleteDocument(ctx context.Context, id int64) error {
	uid, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnDeleteDocument + `($1,$2)`
	_, err := r.db.ExecContext(ctx, query, id, uid)
	return err
}

// SetDocumentPermission -> fn_set_document_permission(admin_id, document_id, target_user_id, can_view, can_edit)
func (r *DocumentPostgres) SetDocumentPermission(ctx context.Context, docID int64, p archive.DocumentPermission) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnSetDocumentPermission + `($1,$2,$3,$4,$5)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, p.UserID, p.CanView, p.CanEdit)
	return err
}

// RemoveDocumentPermission -> fn_remove_document_permission(admin_id, document_id, target_user_id)
func (r *DocumentPostgres) RemoveDocumentPermission(ctx context.Context, docID int64, targetUserID int64) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnRemoveDocumentPermission + `($1,$2,$3)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, targetUserID)
	return err
}
