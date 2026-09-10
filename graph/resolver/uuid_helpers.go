package graph

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func newUUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	}
}

func parseUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID %q: %w", value, err)
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
}

func uuidString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}

	return uuid.UUID(id.Bytes).String()
}
