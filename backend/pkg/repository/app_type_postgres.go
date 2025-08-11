package repository

import (
	"fmt"
	"hotel"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AppTypePostgres struct {
	db *sqlx.DB
}

func NewAppTypePostgres(db *sqlx.DB) *AppTypePostgres {
	return &AppTypePostgres{db: db}
}

func (r *AppTypePostgres) Create(appType hotel.AppType) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createAppTypeQuery := fmt.Sprintf("INSERT INTO %s (app_type_id, app_type_name) VALUES ($1, $2) RETURNING app_type_id", appTypeTable)
	row := tx.QueryRow(createAppTypeQuery,
		appType.App_type_id,
		appType.App_type_name,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *AppTypePostgres) GetAll() ([]hotel.AppType, error) {
	var appTypes []hotel.AppType
	query := fmt.Sprintf("SELECT * FROM %s", appTypeTable)
	err := r.db.Select(&appTypes, query)

	return appTypes, err
}

func (r *AppTypePostgres) GetById(appTypeId int) (hotel.AppType, error) {
	var appType hotel.AppType

	query := fmt.Sprintf(`SELECT * FROM %s tl WHERE tl.app_type_id = $1`,
		appTypeTable)
	err := r.db.Get(&appType, query, appTypeId)
	return appType, err
}

func (r *AppTypePostgres) Delete(appTypeId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.app_type_id = $1`,
		appTypeTable)
	_, err := r.db.Exec(query, appTypeId)

	return err
}

func (r *AppTypePostgres) Update(appTypeId int, input hotel.AppTypeUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.App_type_name != nil {
		setValues = append(setValues, fmt.Sprintf("app_type_name=$%d", argId))
		args = append(args, *input.App_type_name)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.app_type_id = $%d",
		appTypeTable, setQuery, argId)

	args = append(args, appTypeId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
