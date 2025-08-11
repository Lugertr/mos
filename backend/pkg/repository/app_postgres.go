package repository

import (
	"fmt"
	"hotel"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AppPostgres struct {
	db *sqlx.DB
}

func NewAppPostgres(db *sqlx.DB) *AppPostgres {
	return &AppPostgres{db: db}
}

func (r *AppPostgres) Create(app hotel.App) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createAppQuery := fmt.Sprintf("INSERT INTO %s (rooms, app_type_id, app_status, app_price) VALUES ($1, $2, $3, $4) RETURNING app_id", appTable)
	row := tx.QueryRow(createAppQuery,
		app.Rooms,
		app.App_type_id,
		app.AppStatus,
		app.App_price,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *AppPostgres) GetAll() ([]hotel.App, error) {
	var apps []hotel.App
	query := fmt.Sprintf("SELECT * FROM %s", appTable)
	err := r.db.Select(&apps, query)

	return apps, err
}

func (r *AppPostgres) GetById(appId int) ([]hotel.App, error) {
	var app []hotel.App

	query := fmt.Sprintf(`SELECT * FROM %s`,
		appFunc)
	err := r.db.Select(&app, query)
	return app, err
}

func (r *AppPostgres) Delete(appId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.app_id = $1`,
		appTable)
	_, err := r.db.Exec(query, appId)

	return err
}

func (r *AppPostgres) Update(appId int, input hotel.AppUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Rooms != nil {
		setValues = append(setValues, fmt.Sprintf("rooms=$%d", argId))
		args = append(args, *input.Rooms)
		argId++
	}

	if input.App_type_id != nil {
		setValues = append(setValues, fmt.Sprintf("app_type_id=$%d", argId))
		args = append(args, *input.App_type_id)
		argId++
	}

	if input.AppStatus != nil {
		setValues = append(setValues, fmt.Sprintf("app_status=$%d", argId))
		args = append(args, *input.AppStatus)
		argId++
	}

	if input.App_price != nil {
		setValues = append(setValues, fmt.Sprintf("app_price=$%d", argId))
		args = append(args, *input.App_price)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.app_id = $%d",
		appTable, setQuery, argId)

	args = append(args, appId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
