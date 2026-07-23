package guards

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func sameUUID(left, right pgtype.UUID) bool {
	return left.Valid == right.Valid && left.Bytes == right.Bytes
}
