package repository

import (
	"archive"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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

	query := `SELECT fn_add_document($1, $2, $3, $4, $5, $6, $7, $8, $9)`
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

// SearchDocumentsByTag -> fn_get_documents_by_tag_secure(tag, author, type, date_from, date_to, requester_id)
func (r *DocumentPostgres) SearchDocumentsByTag(ctx context.Context, filter archive.DocumentSearchFilter) ([]archive.DocumentSecure, error) {
	query := `
SELECT
  doc_id,
  title,
  privacy,
  created_at,
  created_by,
  created_by_login,
  created_by_full_name,
  updated_at,
  updated_by,
  updated_by_login,
  updated_by_full_name,
  document_date,
  author_id,
  author_name,
  type_id,
  type_name,
  tags,
  viewers,
  editors,
  can_requester_edit,
  ST_AsGeoJSON(geom) as geom
FROM fn_get_documents_by_tag_secure($1,$2,$3,$4,$5,$6)
ORDER BY created_at DESC
`
	var df, dt interface{}
	if strings.TrimSpace(filter.DateFrom) != "" {
		df = filter.DateFrom
	}
	if strings.TrimSpace(filter.DateTo) != "" {
		dt = filter.DateTo
	}

	var requester interface{} = nil
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}

	var out []archive.DocumentSecure
	if err := r.db.SelectContext(ctx, &out, query,
		filter.Tag,
		nullString(filter.Author),
		nullString(filter.Type),
		df,
		dt,
		requester,
	); err != nil {
		return nil, err
	}
	// respect limit/offset on app side if DB func doesn't support — handled in service/handler
	return out, nil
}

func nullString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// GetDocumentByID -> fn_get_document_secure(document_id, requester_id)
func (r *DocumentPostgres) GetDocumentByID(ctx context.Context, id int64) (archive.DocumentSecure, error) {
	var out archive.DocumentSecure
	query := `
SELECT
  id AS doc_id,
  title,
  privacy,
  created_at,
  created_by,
  created_by_login,
  created_by_full_name,
  updated_at,
  updated_by,
  updated_by_login,
  updated_by_full_name,
  document_date,
  author_id,
  author_name,
  type_id,
  type_name,
  tags,
  viewers,
  editors,
  can_requester_edit,
  ST_AsGeoJSON(geom) as geom
FROM fn_get_document_secure($1,$2)
LIMIT 1
`
	var requester interface{} = nil
	if uid, ok := userIDFromCtx(ctx); ok {
		requester = uid
	}
	if err := r.db.GetContext(ctx, &out, query, id, requester); err != nil {
		if err == sql.ErrNoRows {
			return archive.DocumentSecure{}, nil
		}
		return archive.DocumentSecure{}, err
	}
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

	query := `SELECT fn_update_document($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
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
	query := `SELECT fn_delete_document($1,$2)`
	_, err := r.db.ExecContext(ctx, query, id, uid)
	return err
}

// SetDocumentPermission -> fn_set_document_permission(admin_id, document_id, target_user_id, can_view, can_edit)
func (r *DocumentPostgres) SetDocumentPermission(ctx context.Context, docID int64, p archive.DocumentPermission) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT fn_set_document_permission($1,$2,$3,$4,$5)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, p.UserID, p.CanView, p.CanEdit)
	return err
}

// RemoveDocumentPermission -> fn_remove_document_permission(admin_id, document_id, target_user_id)
func (r *DocumentPostgres) RemoveDocumentPermission(ctx context.Context, docID int64, targetUserID int64) error {
	adminID, ok := userIDFromCtx(ctx)
	if !ok {
		return fmt.Errorf("user id missing in context")
	}
	query := `SELECT fn_remove_document_permission($1,$2,$3)`
	_, err := r.db.ExecContext(ctx, query, adminID, docID, targetUserID)
	return err
}
