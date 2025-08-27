package repository

import (
	"context"
	"time"

	"archive"

	"github.com/jmoiron/sqlx"
)

type AdminPostgres struct {
	db *sqlx.DB
}

func NewAdminPostgres(db *sqlx.DB) *AdminPostgres {
	return &AdminPostgres{db: db}
}

func (r *AdminPostgres) GetLogsByUser(ctx context.Context, adminID int64, targetUserID int64, start *time.Time, end *time.Time) ([]archive.LogRecord, error) {
	// SELECT * FROM fn_get_logs_by_user($1,$2,$3,$4)
	query := `SELECT * FROM ` + fnGetLogsByUser + `($1,$2,$3,$4)`
	var s interface{}
	var e interface{}
	if start != nil {
		s = *start
	}
	if end != nil {
		e = *end
	}
	var logs []archive.LogRecord
	if err := r.db.SelectContext(ctx, &logs, query, adminID, targetUserID, s, e); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AdminPostgres) GetLogsByTable(ctx context.Context, adminID int64, tableName string, start *time.Time, end *time.Time) ([]archive.LogRecord, error) {
	// SELECT * FROM fn_get_logs_by_table($1,$2,$3,$4)
	query := `SELECT * FROM ` + fnGetLogsByTable + `($1,$2,$3,$4)`
	var s interface{}
	var e interface{}
	if start != nil {
		s = *start
	}
	if end != nil {
		e = *end
	}
	var logs []archive.LogRecord
	if err := r.db.SelectContext(ctx, &logs, query, adminID, tableName, s, e); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AdminPostgres) GetLogsByDate(ctx context.Context, adminID int64, start time.Time, end time.Time) ([]archive.LogRecord, error) {
	// SELECT * FROM fn_get_logs_by_date($1,$2,$3)
	query := `SELECT * FROM ` + fnGetLogsByDate + `($1,$2,$3)`
	var logs []archive.LogRecord
	if err := r.db.SelectContext(ctx, &logs, query, adminID, start, end); err != nil {
		return nil, err
	}
	return logs, nil
}
