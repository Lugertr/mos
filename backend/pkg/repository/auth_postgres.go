package repository

import (
	"context"
	"database/sql"

	"archive"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (r *AuthPostgres) GetUsersByIDs(ctx context.Context, ids []int64) ([]archive.User, error) {
	const q = `SELECT id, full_name FROM ` + "fn_get_users_by_ids" + `($1)`
	rows, err := r.db.QueryxContext(ctx, q, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]archive.User, 0)
	for rows.Next() {
		var id int64
		var fullName sql.NullString
		if err := rows.Scan(&id, &fullName); err != nil {
			return nil, err
		}
		u := archive.User{
			ID: id,
		}
		if fullName.Valid {
			u.FullName = &fullName.String
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserFullName -> SELECT fn_update_user_full_name(p_requester_id, p_target_user_id, p_full_name)
func (r *AuthPostgres) UpdateUserFullName(ctx context.Context, requesterID int64, targetUserID int64, fullName string) error {
	const q = `SELECT ` + "fn_update_user_full_name" + `($1,$2,$3)`
	_, err := r.db.ExecContext(ctx, q, requesterID, targetUserID, fullName)
	return err
}

// ChangeUserPassword -> SELECT fn_change_user_password(p_requester_id, p_target_user_id, p_old_password, p_new_password)
func (r *AuthPostgres) ChangeUserPassword(ctx context.Context, requesterID int64, targetUserID int64, oldPasswordHash, newPasswordHash string) error {
	const q = `SELECT ` + "fn_change_user_password" + `($1,$2,$3,$4)`
	_, err := r.db.ExecContext(ctx, q, requesterID, targetUserID, oldPasswordHash, newPasswordHash)
	return err
}
