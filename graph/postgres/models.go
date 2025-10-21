package postgres

import (
	"github.com/google/uuid"
)

type DBUser struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}
