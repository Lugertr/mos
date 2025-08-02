package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AnalyzerPostgres struct {
	db *sqlx.DB
}

func NewAnalyzerPostgres(db *sqlx.DB) *AnalyzerPostgres {
	return &AnalyzerPostgres{db: db}
}

func (r *AnalyzerPostgres) Create(analyzer center.Analyzer) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createAnalyzerQuery := fmt.Sprintf("INSERT INTO %s (analyzer_id, order_id, arrival_date_time, completion_date_time, execution_time_in_seconds) VALUES ($1, $2, $3, $4, $5) RETURNING analyzer_id", analyzerTable)
	row := tx.QueryRow(createAnalyzerQuery,
		analyzer.AnalyzerID,
		analyzer.OrderID,
		analyzer.ArrivalDateTime,
		analyzer.CompletionDateTime,
		analyzer.ExecutionTimeInSeconds,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *AnalyzerPostgres) GetAll() ([]center.Analyzer, error) {
	var analyzers []center.Analyzer
	query := fmt.Sprintf("SELECT * FROM %s", analyzerTable)
	err := r.db.Select(&analyzers, query)

	return analyzers, err
}

func (r *AnalyzerPostgres) GetById(analyzerId int) (center.Analyzer, error) {
	var analyzer center.Analyzer

	query := fmt.Sprintf(`SELECT * FROM %s`,
		analyzerTable)
	err := r.db.Get(&analyzer, query, analyzerId)
	return analyzer, err
}

func (r *AnalyzerPostgres) Delete(analyzerId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.analyzer_id = $1`,
		analyzerTable)
	_, err := r.db.Exec(query, analyzerId)

	return err
}

func (r *AnalyzerPostgres) Update(analyzerId int, input center.AnalyzerUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.OrderID != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.OrderID)
		argId++
	}

	if input.ArrivalDateTime != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.ArrivalDateTime)
		argId++
	}

	if input.CompletionDateTime != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.CompletionDateTime)
		argId++
	}

	if input.ExecutionTimeInSeconds != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.ExecutionTimeInSeconds)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.analyzer_id=$%d",
		analyzerTable, setQuery, argId)

	args = append(args, analyzerId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
