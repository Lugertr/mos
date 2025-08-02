package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type PatientPostgres struct {
	db *sqlx.DB
}

func NewPatientPostgres(db *sqlx.DB) *PatientPostgres {
	return &PatientPostgres{db: db}
}

func (r *PatientPostgres) Create(providedService center.Patient) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createPatientQuery := fmt.Sprintf("INSERT INTO %s (patient_id, full_name,date_of_birth,passport_serial_number,phone,email,insurance_number,insurance_type,insurance_company) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING patient_id", patientTable)

	row := tx.QueryRow(createPatientQuery,
		providedService.PatientID,
		providedService.FullName,
		providedService.DateOfBirth,
		providedService.PassportSerialNumber,
		providedService.Phone,
		providedService.Email,
		providedService.InsuranceNumber,
		providedService.InsuranceType,
		providedService.InsuranceCompany,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *PatientPostgres) GetAll() ([]center.Patient, error) {
	var providedServices []center.Patient
	query := fmt.Sprintf("SELECT * FROM %s", patientTable)
	err := r.db.Select(&providedServices, query)

	return providedServices, err
}

func (r *PatientPostgres) GetById(providedServiceId int) (center.Patient, error) {
	var patient center.Patient

	query := fmt.Sprintf(`SELECT * FROM %s`,
		patientTable)
	err := r.db.Get(&patient, query, providedServiceId)
	return patient, err
}

func (r *PatientPostgres) Delete(providedServiceId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.providedService_id = $1`,
		patientTable)
	_, err := r.db.Exec(query, providedServiceId)

	return err
}

func (r *PatientPostgres) Update(providedServiceId int, input center.PatientUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.FullName != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.FullName)
		argId++
	}

	if input.DateOfBirth != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.DateOfBirth)
		argId++
	}

	if input.PassportSerialNumber != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.PassportSerialNumber)
		argId++
	}

	if input.Phone != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.Phone)
		argId++
	}

	if input.Email != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.Email)
		argId++
	}

	if input.InsuranceNumber != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.InsuranceNumber)
		argId++
	}

	if input.InsuranceType != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.InsuranceType)
		argId++
	}

	if input.InsuranceCompany != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.InsuranceCompany)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.providedService_id=$%d",
		patientTable, setQuery, argId)

	args = append(args, providedServiceId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
