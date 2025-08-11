package repository

import (
	"fmt"
	"hotel"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type ClientPostgres struct {
	db *sqlx.DB
}

func NewClientPostgres(db *sqlx.DB) *ClientPostgres {
	return &ClientPostgres{db: db}
}

func (r *ClientPostgres) Create(client hotel.Client) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createclientQuery := fmt.Sprintf("INSERT INTO %s (client_name,family_name,surname,passport,gender,app_id,date_in,date_out) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING client_id", clientTable)
	row := tx.QueryRow(createclientQuery,
		client.Client_name,
		client.Family_name,
		client.Surname,
		client.Passport,
		client.Gender,
		client.App_id,
		client.Date_in,
		client.Date_out,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *ClientPostgres) GetAll() ([]hotel.Client, error) {
	var clients []hotel.Client
	query := fmt.Sprintf("SELECT * FROM %s", clientTable)
	err := r.db.Select(&clients, query)

	return clients, err
}

func (r *ClientPostgres) GetById(clientId int) (hotel.ClientFunc, error) {
	var client hotel.ClientFunc

	query := fmt.Sprintf(`SELECT * FROM %s`,
		clientFunc)
	err := r.db.Get(&client, query, clientId)
	return client, err
}

func (r *ClientPostgres) Delete(clientId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.client_id = $1`,
		clientTable)
	_, err := r.db.Exec(query, clientId)

	return err
}

func (r *ClientPostgres) Update(clientId int, input hotel.ClientUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Client_name != nil {
		setValues = append(setValues, fmt.Sprintf("client_name=$%d", argId))
		args = append(args, *input.Client_name)
		argId++
	}

	if input.Family_name != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.Family_name)
		argId++
	}

	if input.Surname != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.Surname)
		argId++
	}

	if input.Passport != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.Passport)
		argId++
	}

	if input.Gender != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.Gender)
		argId++
	}

	if input.App_id != nil {
		setValues = append(setValues, fmt.Sprintf("app_id=$%d", argId))
		args = append(args, *input.App_id)
		argId++
	}

	if input.Date_in != nil {
		setValues = append(setValues, fmt.Sprintf("date_in=$%d", argId))
		args = append(args, *input.Date_in)
		argId++
	}

	if input.Date_out != nil {
		setValues = append(setValues, fmt.Sprintf("date_out=$%d", argId))
		args = append(args, *input.Date_out)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.client_id=$%d",
		clientTable, setQuery, argId)

	args = append(args, clientId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
