package repository

import (
	"context"
	"fmt"

	"archive"

	"github.com/jmoiron/sqlx"
)

type DocumentTypesPostgres struct {
	db *sqlx.DB
}

func NewDocumentTypesPostgres(db *sqlx.DB) *DocumentTypesPostgres {
	return &DocumentTypesPostgres{db: db}
}

func (r *DocumentTypesPostgres) CreateDocumentType(ctx context.Context, t archive.DocumentType) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	ins := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, documentTypesTable)
	if _, err := tx.ExecContext(ctx, ins, t.Name); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	var id int64
	sel := fmt.Sprintf(`SELECT id FROM %s WHERE name = $1`, documentTypesTable)
	if err := tx.GetContext(ctx, &id, sel, t.Name); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *DocumentTypesPostgres) GetAllDocumentTypes(ctx context.Context) ([]archive.DocumentType, error) {
	var out []archive.DocumentType
	q := fmt.Sprintf(`SELECT id, name FROM %s ORDER BY lower(name)`, documentTypesTable)
	if err := r.db.SelectContext(ctx, &out, q); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *DocumentTypesPostgres) GetDocumentType(ctx context.Context, id int64) (archive.DocumentType, error) {
	var t archive.DocumentType
	q := fmt.Sprintf(`SELECT id, name FROM %s WHERE id = $1`, documentTypesTable)
	err := r.db.GetContext(ctx, &t, q, id)
	return t, err
}

func (r *DocumentTypesPostgres) UpdateDocumentType(ctx context.Context, id int64, t archive.DocumentType) error {
	q := fmt.Sprintf(`UPDATE %s SET name = $1 WHERE id = $2`, documentTypesTable)
	_, err := r.db.ExecContext(ctx, q, t.Name, id)
	return err
}

func (r *DocumentTypesPostgres) DeleteDocumentType(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, documentTypesTable)
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
