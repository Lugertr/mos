package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type LabServicePostgres struct {
	db *sqlx.DB
}

func NewLabServicePostgres(db *sqlx.DB) *LabServicePostgres {
	return &LabServicePostgres{db: db}
}

func (r *LabServicePostgres) Create(labService center.LabService) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createLabServiceQuery := fmt.Sprintf("INSERT INTO %s (service_id, name, cost, service_code, execution_time, average_deviation) VALUES ($1, $2, $3, $4, $5, $6) RETURNING service_id", labServiceTable)
	row := tx.QueryRow(createLabServiceQuery,
		labService.ServiceID,
		labService.Name,
		labService.Cost,
		labService.ServiceCode,
		labService.ExecutionTime,
		labService.AverageDeviation,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *LabServicePostgres) GetAll() ([]center.LabService, error) {
	var labServices []center.LabService
	query := fmt.Sprintf("SELECT * FROM %s", labServiceTable)
	err := r.db.Select(&labServices, query)

	return labServices, err
}

func (r *LabServicePostgres) GetById(labServiceId int) (center.LabService, error) {
	var labService center.LabService

	query := fmt.Sprintf(`SELECT * FROM %s`,
		labServiceTable)
	err := r.db.Get(&labService, query, labServiceId)
	return labService, err
}

func (r *LabServicePostgres) Delete(labServiceId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.labService_id = $1`,
		labServiceTable)
	_, err := r.db.Exec(query, labServiceId)

	return err
}

func (r *LabServicePostgres) Update(labServiceId int, input center.LabServiceUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Name != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.Name)
		argId++
	}

	if input.Cost != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.Cost)
		argId++
	}

	if input.ServiceCode != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.ServiceCode)
		argId++
	}

	if input.ExecutionTime != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.ExecutionTime)
		argId++
	}

	if input.AverageDeviation != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.AverageDeviation)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.labService_id=$%d",
		labServiceTable, setQuery, argId)

	args = append(args, labServiceId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
