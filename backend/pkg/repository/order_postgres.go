package repository

import (
	"center"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

func (r *OrderPostgres) Create(order center.Order) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createOrderQuery := fmt.Sprintf("INSERT INTO %s (service_id, name, cost, service_code, execution_time, average_deviation) VALUES ($1, $2, $3, $4, $5, $6) RETURNING service_id", orderTable)
	row := tx.QueryRow(createOrderQuery,
		order.OrderID,
		order.CreationDate,
		order.PatientID,
		order.StatusOrder,
		order.ExecutionTimeInDays,
	)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *OrderPostgres) GetAll() ([]center.Order, error) {
	var orders []center.Order
	query := fmt.Sprintf("SELECT * FROM %s", orderTable)
	err := r.db.Select(&orders, query)

	return orders, err
}

func (r *OrderPostgres) GetById(orderId int) (center.Order, error) {
	var order center.Order

	query := fmt.Sprintf(`SELECT * FROM %s`,
		orderTable)
	err := r.db.Get(&order, query, orderId)
	return order, err
}

func (r *OrderPostgres) Delete(orderId int) error {
	query := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.order_id = $1`,
		orderTable)
	_, err := r.db.Exec(query, orderId)

	return err
}

func (r *OrderPostgres) Update(orderId int, input center.OrderUpdate) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.CreationDate != nil {
		setValues = append(setValues, fmt.Sprintf("family_name=$%d", argId))
		args = append(args, *input.CreationDate)
		argId++
	}

	if input.PatientID != nil {
		setValues = append(setValues, fmt.Sprintf("surname=$%d", argId))
		args = append(args, *input.PatientID)
		argId++
	}

	if input.StatusOrder != nil {
		setValues = append(setValues, fmt.Sprintf("passport=$%d", argId))
		args = append(args, *input.StatusOrder)
		argId++
	}

	if input.ExecutionTimeInDays != nil {
		setValues = append(setValues, fmt.Sprintf("gender=$%d", argId))
		args = append(args, *input.ExecutionTimeInDays)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.order_id=$%d",
		orderTable, setQuery, argId)

	args = append(args, orderId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
