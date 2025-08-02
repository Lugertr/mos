package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type ProvidedServicePostgres struct {
	db *sqlx.DB
}

func NewProvidedServicePostgres(db *sqlx.DB) *ProvidedServicePostgres {
	return &ProvidedServicePostgres{db: db}
}

func (r *ProvidedServicePostgres) Create(providedService center.ProvidedService) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createProvidedServiceQuery := fmt.Sprintf("INSERT INTO %s provided_services (provided_service_id, service_id, order_id, execution_date, performer) VALUES ($1, $2, $3, $4, $5) RETURNING provided_service_id", providedServiceTable)
	row := tx.QueryRow(createProvidedServiceQuery,
		providedService.ProvidedServiceID,
		providedService.ServiceID,
		providedService.OrderID,
		providedService.ExecutionDate,
		providedService.Performer,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *ProvidedServicePostgres) GetAll() ([]center.ProvidedService, error) {
	var providedServices []center.ProvidedService
	query := fmt.Sprintf("SELECT * FROM %s", providedServiceTable)
	err := r.db.Select(&providedServices, query)

	return providedServices, err
}

func (r *ProvidedServicePostgres) GetById(providedServiceId int) (center.ProvidedService, error) {
	var providedService center.ProvidedService

	query := fmt.Sprintf(`SELECT * FROM %s`,
		providedServiceTable)
	err := r.db.Get(&providedService, query, providedServiceId)
	return providedService, err
}

func (r *ProvidedServicePostgres) Delete(providedServiceId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.providedService_id = $1`,
		providedServiceTable)
	_, err := r.db.Exec(query, providedServiceId)

	return err
}

func (r *ProvidedServicePostgres) Update(providedServiceId int, input center.ProvidedServiceUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.ServiceID != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.ServiceID)
		argId++
	}

	if input.OrderID != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.OrderID)
		argId++
	}

	if input.ExecutionDate != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.ExecutionDate)
		argId++
	}

	if input.Performer != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.Performer)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.providedService_id=$%d",
		providedServiceTable, setQuery, argId)

	args = append(args, providedServiceId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
