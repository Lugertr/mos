package repository

import (
	"center"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user center.UserCreate) (int, error) {
	var (
		id    int
		query string
		row   *sql.Row
	)

	query = fmt.Sprintf("INSERT INTO %s (username, password_hash, user_type) values ($1, $2, $3) RETURNING id", usersTable)
	row = r.db.QueryRow(query, user.Username, user.Password, user.UserType)

	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUser(username, password string) (center.UserRet, error) {
	logrus.Print(password)
	var user center.UserRet
	query := fmt.Sprintf("SELECT id FROM %s WHERE username=$1 AND password_hash=$2", usersTable)
	err := r.db.Get(&user, query, username, password)

	return user, err
}

func (r *AuthPostgres) CheckUser(userId string) (bool, error) {
	var userID int
	query := fmt.Sprintf("SELECT id FROM %s WHERE user_id=$1", usersTable)
	err := r.db.Get(&userID, query, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
