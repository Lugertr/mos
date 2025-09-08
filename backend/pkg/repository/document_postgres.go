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

// userIDFromCtx возвращает user id из контекста (поддерживает int/int64/*int64)
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

// --- helpers for nullable params ---
// возвращает nil если geojson == nil или пустой, иначе валидированный json.RawMessage
func geoJSONParam(m *json.RawMessage) (interface{}, error) {
	if m == nil || len(*m) == 0 {
		return nil, nil
	}
	if !json.Valid(*m) {
		return nil, fmt.Errorf("invalid geojson")
	}
	return *m, nil
}

func bytesParam(b *[]byte) interface{} {
	if b == nil || len(*b) == 0 {
		// если хотим отличать пустой слайс от NULL, можно вернуть *b,
		// но текущее поведение: пустой слайс -> NULL (как раньше).
		return nil
	}
	return *b
}

func trimStringParam(s *string) interface{} {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return t
}

func privacyParam(p archive.PrivacyType) interface{} {
	if p == "" {
		return nil
	}
	return string(p)
}

// --- CreateDocument -> fn_add_document ---
func (r *DocumentPostgres) CreateDocument(ctx context.Context, in archive.DocumentCreateInput) (int64, error) {
	var id int64

	geojsonVal, err := geoJSONParam(in.GeoJSON)
	if err != nil {
		return 0, err
	}
	fileVal := bytesParam(in.File)
	authorVal := trimStringParam(in.Author)
	privacyVal := privacyParam(in.Privacy)

	query := `SELECT ` + fnAddDocument + `($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	err = r.db.QueryRowxContext(ctx, query,
		in.CreatorID,
		in.Title,
		in.DocumentDate,
		authorVal,
		in.TypeID,
		fileVal,
		geojsonVal,
		in.Tags,
		privacyVal,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// --- SearchDocumentsByTag -> fn_get_documents_for_user ---
func (r *DocumentPostgres) SearchDocumentsByTag(ctx context.Context, filter archive.DocumentSearchFilter) ([]archive.DocumentSecure, error) {
	const q = `
SELECT
  id,
  title,
  privacy,
  updated_at,
  document_date,
  type_id,
  author,
  geojson,
  can_edit,
  is_author
FROM ` + fnGetDocumentsForUser + `($1)
ORDER BY COALESCE(updated_at, now()) DESC
`

	var requester interface{}
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}

	type listRow struct {
		ID           int64               `db:"id"`
		Title        string              `db:"title"`
		Privacy      archive.PrivacyType `db:"privacy"`
		UpdatedAt    sql.NullTime        `db:"updated_at"`
		DocumentDate *time.Time          `db:"document_date"`
		TypeID       *int64              `db:"type_id"`
		Author       sql.NullString      `db:"author"`
		GeoJSON      *json.RawMessage    `db:"geojson"`
		CanEdit      bool                `db:"can_edit"`
		IsAuthor     bool                `db:"is_author"`
	}

	var rows []listRow
	if err := r.db.SelectContext(ctx, &rows, q, requester); err != nil {
		return nil, err
	}

	out := make([]archive.DocumentSecure, 0, len(rows))
	for _, rr := range rows {
		var updatedAtPtr *time.Time
		if rr.UpdatedAt.Valid {
			t := rr.UpdatedAt.Time
			updatedAtPtr = &t
		}
		var authorPtr *string
		if rr.Author.Valid {
			s := rr.Author.String
			authorPtr = &s
		}
		ds := archive.DocumentSecure{
			DocID:            rr.ID,
			Title:            rr.Title,
			Privacy:          rr.Privacy,
			UpdatedAt:        updatedAtPtr,
			DocumentDate:     rr.DocumentDate,
			Author:           authorPtr,
			TypeID:           rr.TypeID,
			Tags:             nil,
			Viewers:          nil,
			Editors:          nil,
			CanRequesterEdit: rr.CanEdit,
		}
		if rr.GeoJSON != nil && len(*rr.GeoJSON) > 0 {
			s := string(*rr.GeoJSON)
			ds.Geom = &s
		}
		out = append(out, ds)
	}

	// apply offset/limit in app layer
	start := filter.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(out) {
		return []archive.DocumentSecure{}, nil
	}
	end := len(out)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return out[start:end], nil
}

func nullString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// --- GetDocumentByID -> fn_get_document_by_id ---
func (r *DocumentPostgres) GetDocumentByID(ctx context.Context, id int64) (archive.DocumentSecure, error) {
	const q = `
SELECT
  id,
  title,
  privacy,
  created_at,
  created_by,
  updated_at,
  updated_by,
  document_date,
  author,
  type_id,
  file_bytea,
  geojson,
  ST_AsGeoJSON(geom) as geom,
  can_edit
FROM ` + fnGetDocumentByID + `($1,$2)
LIMIT 1
`

	var requester interface{}
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}

	type docRow struct {
		ID           int64               `db:"id"`
		Title        string              `db:"title"`
		Privacy      archive.PrivacyType `db:"privacy"`
		CreatedAt    time.Time           `db:"created_at"`
		CreatedBy    *int64              `db:"created_by"`
		UpdatedAt    sql.NullTime        `db:"updated_at"`
		UpdatedBy    *int64              `db:"updated_by"`
		DocumentDate *time.Time          `db:"document_date"`
		Author       sql.NullString      `db:"author"`
		TypeID       *int64              `db:"type_id"`
		File         []byte              `db:"file_bytea"`
		GeoJSON      *json.RawMessage    `db:"geojson"`
		Geom         *string             `db:"geom"`
		CanEdit      bool                `db:"can_edit"`
	}

	var row docRow
	if err := r.db.GetContext(ctx, &row, q, id, requester); err != nil {
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
	var authorPtr *string
	if row.Author.Valid {
		s := row.Author.String
		authorPtr = &s
	}

	out := archive.DocumentSecure{
		DocID:            row.ID,
		Title:            row.Title,
		Privacy:          row.Privacy,
		CreatedAt:        row.CreatedAt,
		CreatedBy:        row.CreatedBy,
		UpdatedAt:        updatedAtPtr,
		UpdatedBy:        row.UpdatedBy,
		DocumentDate:     row.DocumentDate,
		Author:           authorPtr,
		TypeID:           row.TypeID,
		Tags:             nil,
		Viewers:          nil,
		Editors:          nil,
		CanRequesterEdit: row.CanEdit,
		Geom:             row.Geom,
	}

	_ = row.File
	_ = row.GeoJSON

	return out, nil
}

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

	titleParam := interface{}(nil)
	if in.Title != nil {
		titleParam = *in.Title
	}

	fileParam := bytesParam(in.File)

	geojsonVal, err := geoJSONParam(in.GeoJSON)
	if err != nil {
		return err
	}

	tagsParam := interface{}(nil)
	if in.Tags != nil {
		tagsParam = *in.Tags
	}

	privacyVal := interface{}(nil)
	if in.Privacy != nil {
		privacyVal = string(*in.Privacy)
	}

	authorVal := trimStringParam(in.Author)

	query := `SELECT ` + fnUpdateDocument + `($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err = r.db.ExecContext(ctx, query,
		in.DocumentID,
		in.UpdaterID,
		titleParam,
		in.DocumentDate,
		authorVal,
		in.TypeID,
		fileParam,
		geojsonVal,
		tagsParam,
		privacyVal,
	)
	return err
}

// DeleteDocument, SetDocumentPermission, RemoveDocumentPermission — без изменений
func (r *DocumentPostgres) DeleteDocument(ctx context.Context, id int64) error {
	uid, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnDeleteDocument + `($1,$2)`
	_, err := r.db.ExecContext(ctx, query, id, uid)
	return err
}

func (r *DocumentPostgres) SetDocumentPermission(ctx context.Context, docID int64, p archive.DocumentPermission) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnSetDocumentPermission + `($1,$2,$3,$4,$5)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, p.UserID, p.CanView, p.CanEdit)
	return err
}

func (r *DocumentPostgres) RemoveDocumentPermission(ctx context.Context, docID int64, targetUserID int64) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT ` + fnRemoveDocumentPermission + `($1,$2,$3)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, targetUserID)
	return err
}
