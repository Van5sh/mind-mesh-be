package services

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

// fetchGrouped backs the "list of X per parent" batch methods the
// dataloaders are built on (GetXByProjectIDs, GetXByChatIDs, ...): it
// validates the IDs, runs the single batched query, and groups the rows by
// their parent ID. Every requested ID is present in the result (nil slice
// when that parent has no rows), so callers never see a missing key for a
// parent that simply has nothing yet.
//
// what is the full error text ("failed to get ..." is added), keyOf picks the
// parent ID out of a row.
func fetchGrouped[T any](
	ctx context.Context,
	ids []pgtype.UUID,
	what string,
	fetch func(ctx context.Context, ids []pgtype.UUID) ([]T, error),
	keyOf func(T) pgtype.UUID,
) (map[pgtype.UUID][]T, error) {
	if err := validateIDs(ids); err != nil {
		return nil, err
	}

	rows, err := fetch(ctx, ids)
	if err != nil {
		return nil, apperrors.InternalError("failed to get "+what, err)
	}

	grouped := make(map[pgtype.UUID][]T, len(ids))
	for _, id := range ids {
		grouped[id] = nil
	}
	for _, row := range rows {
		id := keyOf(row)
		grouped[id] = append(grouped[id], row)
	}

	return grouped, nil
}

// fetchRows backs the one-row-per-ID batch methods (profiles, metadata,
// storage, folders by ID): validate, run the single batched query. IDs with no
// row are simply absent from the result - the loader turns that into
// "not found".
func fetchRows[T any](
	ctx context.Context,
	ids []pgtype.UUID,
	what string,
	fetch func(ctx context.Context, ids []pgtype.UUID) ([]T, error),
) ([]T, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if err := validateIDs(ids); err != nil {
		return nil, err
	}

	rows, err := fetch(ctx, ids)
	if err != nil {
		return nil, apperrors.InternalError("failed to get "+what, err)
	}
	return rows, nil
}

func validateIDs(ids []pgtype.UUID) error {
	for _, id := range ids {
		if err := validators.ValidateUUID("id", id); err != nil {
			return err
		}
	}
	return nil
}
