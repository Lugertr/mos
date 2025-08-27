package repository

import (
	"context"

	"archive"

	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

// CreateUser -> вызывает fn_register_user(login, password, full_name)
func (r *AuthPostgres) CreateUser(ctx context.Context, user archive.User) (int64, error) {
	var id int64
	// fn_register_user возвращает integer (new id)
	query := `SELECT ` + fnRegisterUser + `($1, $2, $3)`
	err := r.db.QueryRowxContext(ctx, query, user.Login, user.PasswordHash, user.FullName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUser -> вызывает fn_authorize_user(login, password)
// Примечание: fn_authorize_user в вашем SQL сравнивает значения password напрямую,
// поэтому caller должен договариваться каким форматом пароля пользоваться (plain vs hash).
func (r *AuthPostgres) GetUser(ctx context.Context, login, passwordHash string) (archive.User, error) {
	var u archive.User
	// fn_authorize_user возвращает (id, login, full_name, role_name)
	query := `SELECT id, login, full_name FROM ` + fnAuthorizeUser + `($1, $2)`
	// используем GetContext, чтобы завести результат в struct
	// но fn_authorize_user - это setof record, поэтому лучше использовать QueryRowxContext + Scan
	row := r.db.QueryRowxContext(ctx, query, login, passwordHash)

	var id int64
	var lg string
	var fullName *string
	if err := row.Scan(&id, &lg, &fullName); err != nil {
		return archive.User{}, err
	}
	u.ID = id
	u.Login = lg
	if fullName != nil {
		u.FullName = fullName
	}
	return u, nil
}
