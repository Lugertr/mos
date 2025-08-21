package archive

import (
	"time"
)

type Role struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type User struct {
	ID           int64     `db:"id" json:"id"`
	RoleID       int64     `db:"role_id" json:"role_id"`
	Login        string    `db:"login" json:"login"`
	PasswordHash string    `db:"password_hash" json:"-"`
	FullName     *string   `db:"full_name" json:"full_name,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
