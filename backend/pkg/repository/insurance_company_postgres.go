package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type InsuranceCompanyPostgres struct {
	db *sqlx.DB
}

func NewInsuranceCompanyPostgres(db *sqlx.DB) *InsuranceCompanyPostgres {
	return &InsuranceCompanyPostgres{db: db}
}

func (r *InsuranceCompanyPostgres) Create(InsuranceCompany center.InsuranceCompany) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createInsuranceCompanyQuery := fmt.Sprintf("INSERT INTO %s (insurance_company_id, name, address, inn, bank_account, bik) VALUES ($1, $2, $3, $4, $5, $6) RETURNING insurance_company_id", insuranceCompanyTable)
	row := tx.QueryRow(createInsuranceCompanyQuery,
		InsuranceCompany.InsuranceCompanyID,
		InsuranceCompany.Name,
		InsuranceCompany.Address,
		InsuranceCompany.INN,
		InsuranceCompany.BankAccount,
		InsuranceCompany.BIK,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *InsuranceCompanyPostgres) GetAll() ([]center.InsuranceCompany, error) {
	var InsuranceCompanys []center.InsuranceCompany
	query := fmt.Sprintf("SELECT * FROM %s", insuranceCompanyTable)
	err := r.db.Select(&InsuranceCompanys, query)

	return InsuranceCompanys, err
}

func (r *InsuranceCompanyPostgres) GetById(InsuranceCompanyId int) (center.InsuranceCompany, error) {
	var InsuranceCompany center.InsuranceCompany

	query := fmt.Sprintf(`SELECT * FROM %s`,
		insuranceCompanyTable)
	err := r.db.Get(&InsuranceCompany, query, InsuranceCompanyId)
	return InsuranceCompany, err
}

func (r *InsuranceCompanyPostgres) Delete(InsuranceCompanyId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.InsuranceCompany_id = $1`,
		insuranceCompanyTable)
	_, err := r.db.Exec(query, InsuranceCompanyId)

	return err
}

func (r *InsuranceCompanyPostgres) Update(InsuranceCompanyId int, input center.InsuranceCompanyUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Name != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.Name)
		argId++
	}

	if input.Address != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.Address)
		argId++
	}

	if input.INN != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.INN)
		argId++
	}

	if input.BankAccount != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.BankAccount)
		argId++
	}

	if input.BIK != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.BIK)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.InsuranceCompany_id=$%d",
		insuranceCompanyTable, setQuery, argId)

	args = append(args, InsuranceCompanyId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
