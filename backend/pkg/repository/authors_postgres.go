package repository

import (
	"context"
	"fmt"

	"archive"

	"github.com/jmoiron/sqlx"
)

type AuthorsPostgres struct {
	db *sqlx.DB
}

func NewAuthorsPostgres(db *sqlx.DB) *AuthorsPostgres {
	return &AuthorsPostgres{db: db}
}

func (r *AuthorsPostgres) CreateAuthor(ctx context.Context, a archive.Author) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	ins := fmt.Sprintf(`INSERT INTO %s (full_name) VALUES ($1) ON CONFLICT (full_name) DO NOTHING`, authorsTable)
	if _, err := tx.ExecContext(ctx, ins, a.FullName); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	var id int64
	sel := fmt.Sprintf(`SELECT id FROM %s WHERE full_name = $1`, authorsTable)
	if err := tx.GetContext(ctx, &id, sel, a.FullName); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthorsPostgres) GetAllAuthors(ctx context.Context) ([]archive.Author, error) {
	var out []archive.Author
	q := fmt.Sprintf(`SELECT id, full_name FROM %s ORDER BY lower(full_name)`, authorsTable)
	if err := r.db.SelectContext(ctx, &out, q); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AuthorsPostgres) GetAuthor(ctx context.Context, id int64) (archive.Author, error) {
	var a archive.Author
	q := fmt.Sprintf(`SELECT id, full_name FROM %s WHERE id = $1`, authorsTable)
	err := r.db.GetContext(ctx, &a, q, id)
	return a, err
}

func (r *AuthorsPostgres) UpdateAuthor(ctx context.Context, id int64, a archive.Author) error {
	q := fmt.Sprintf(`UPDATE %s SET full_name = $1 WHERE id = $2`, authorsTable)
	_, err := r.db.ExecContext(ctx, q, a.FullName, id)
	return err
}

func (r *AuthorsPostgres) DeleteAuthor(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, authorsTable)
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
