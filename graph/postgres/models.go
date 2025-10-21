package postgres

import (
	"github.com/google/uuid"
)

type DBUser struct {
	tableName struct{}  `pg:"users"`
	ID        uuid.UUID `pg:"type:uuid,pk"`
	Name      string
	Email     string
	Password  string
}
