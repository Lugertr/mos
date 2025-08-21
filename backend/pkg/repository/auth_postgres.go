package repository

import (
	"context"
	"fmt"

	"archive"

	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(ctx context.Context, user archive.User) (int64, error) {
	var id int64
	query := fmt.Sprintf(`INSERT INTO %s (role_id, login, password_hash, full_name, created_at) VALUES ($1, $2, $3, $4, now()) RETURNING id`, usersTable)
	err := r.db.QueryRowxContext(ctx, query, user.RoleID, user.Login, user.PasswordHash, user.FullName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUser(ctx context.Context, login, passwordHash string) (archive.User, error) {
	var u archive.User
	query := fmt.Sprintf(`SELECT id, role_id, login, password_hash, full_name, created_at FROM %s WHERE login = $1 AND password_hash = $2`, usersTable)
	err := r.db.GetContext(ctx, &u, query, login, passwordHash)
	return u, err
}
