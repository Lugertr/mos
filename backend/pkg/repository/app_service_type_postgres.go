package repository

import (
	"fmt"
	"hotel"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AppServiceTypePostgres struct {
	db *sqlx.DB
}

func NewAppServiceTypePostgres(db *sqlx.DB) *AppServiceTypePostgres {
	return &AppServiceTypePostgres{db: db}
}

func (r *AppServiceTypePostgres) Create(serviceType hotel.AppServiceType) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createclientQuery := fmt.Sprintf("INSERT INTO %s (service_type_name, price) VALUES ($1, $2) RETURNING service_type_id", appServiceTypeTable)
	row := tx.QueryRow(createclientQuery,
		serviceType.Service_type_name,
		serviceType.Price,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *AppServiceTypePostgres) GetAll() ([]hotel.AppServiceType, error) {
	var serviceType []hotel.AppServiceType
	query := fmt.Sprintf("SELECT * FROM %s", appServiceTypeTable)
	err := r.db.Select(&serviceType, query)

	return serviceType, err
}

func (r *AppServiceTypePostgres) GetById(serviceTypeId int) (hotel.AppServiceType, error) {
	var serviceType hotel.AppServiceType

	query := fmt.Sprintf(`SELECT * FROM %s tl WHERE tl.service_type_id = $1`,
		appServiceTypeTable)
	err := r.db.Get(&serviceType, query, serviceTypeId)
	return serviceType, err
}

func (r *AppServiceTypePostgres) Delete(serviceTypeId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.service_type_id = $1`,
		appServiceTypeTable)
	_, err := r.db.Exec(query, serviceTypeId)

	return err
}

func (r *AppServiceTypePostgres) Update(serviceId int, input hotel.AppServiceTypeUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Service_type_name != nil {
		setValues = append(setValues, fmt.Sprintf("service_type_name=$%d", argId))
		args = append(args, *input.Service_type_name)
		argId++
	}

	if input.Price != nil {
		setValues = append(setValues, fmt.Sprintf("Price=$%d", argId))
		args = append(args, *input.Price)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.service_type_id = $%d",
		appServiceTypeTable, setQuery, argId)

	args = append(args, serviceId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
