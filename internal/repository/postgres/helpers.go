package postgres

import (
	"database/sql"
	"errors"
	"time"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func nullTimeUTCPtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time.UTC()
	return &t
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}
