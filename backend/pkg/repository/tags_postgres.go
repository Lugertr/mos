package repository

import (
	"context"
	"fmt"
	"strings"

	"archive"

	"github.com/jmoiron/sqlx"
)

type TagsPostgres struct {
	db *sqlx.DB
}

func NewTagsPostgres(db *sqlx.DB) *TagsPostgres {
	return &TagsPostgres{db: db}
}

func (r *TagsPostgres) CreateTag(ctx context.Context, t archive.Tag) (int64, error) {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return 0, fmt.Errorf("tag name is required")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	ins := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, tagsTable)
	if _, err := tx.ExecContext(ctx, ins, name); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	var id int64
	sel := fmt.Sprintf(`SELECT id FROM %s WHERE name = $1`, tagsTable)
	if err := tx.GetContext(ctx, &id, sel, name); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TagsPostgres) GetAllTags(ctx context.Context) ([]archive.Tag, error) {
	var out []archive.Tag
	q := fmt.Sprintf(`SELECT id, name FROM %s ORDER BY lower(name)`, tagsTable)
	if err := r.db.SelectContext(ctx, &out, q); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TagsPostgres) GetTag(ctx context.Context, id int64) (archive.Tag, error) {
	var t archive.Tag
	q := fmt.Sprintf(`SELECT id, name FROM %s WHERE id = $1`, tagsTable)
	if err := r.db.GetContext(ctx, &t, q, id); err != nil {
		return archive.Tag{}, err
	}
	return t, nil
}

func (r *TagsPostgres) UpdateTag(ctx context.Context, id int64, t archive.Tag) error {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return fmt.Errorf("tag name is required")
	}
	q := fmt.Sprintf(`UPDATE %s SET name = $1 WHERE id = $2`, tagsTable)
	_, err := r.db.ExecContext(ctx, q, name, id)
	return err
}

func (r *TagsPostgres) DeleteTag(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, tagsTable)
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
