package repository

import (
	"context"

	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type OAuthRepository struct {
	queries *database.Queries
}

func NewOAuthRepository(queries *database.Queries) *OAuthRepository {
	return &OAuthRepository{
		queries: queries,
	}
}

// GetOAuthAccount finds an OAuth account using the provider
// and the user's ID from that provider.
func (r *OAuthRepository) GetOAuthAccount(
	ctx context.Context,
	provider string,
	providerUserID string,
) (database.GetOAuthAccountRow, error) {
	return r.queries.GetOAuthAccount(ctx, database.GetOAuthAccountParams{
		Provider:       provider,
		ProviderUserID: providerUserID,
	})
}

// GetOAuthAccountByUserAndProvider finds an OAuth account
// belonging to a specific user and provider.
func (r *OAuthRepository) GetOAuthAccountByUserAndProvider(
	ctx context.Context,
	userID pgtype.UUID,
	provider string,
) (database.OauthAccount, error) {
	return r.queries.GetOAuthAccountByUserAndProvider(
		ctx,
		database.GetOAuthAccountByUserAndProviderParams{
			UserID:   userID,
			Provider: provider,
		},
	)
}

// GetOAuthAccountsByUserID returns all OAuth accounts
// linked to a user.
func (r *OAuthRepository) GetOAuthAccountsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.OauthAccount, error) {
	return r.queries.GetOAuthAccountsByUserID(ctx, userID)
}

// CreateOAuthAccount links an OAuth provider account
// to an existing application user.
func (r *OAuthRepository) CreateOAuthAccount(
	ctx context.Context,
	userID pgtype.UUID,
	provider string,
	providerUserID string,
) (database.OauthAccount, error) {
	return r.queries.CreateOAuthAccount(
		ctx,
		database.CreateOAuthAccountParams{
			UserID:         userID,
			Provider:       provider,
			ProviderUserID: providerUserID,
		},
	)
}

// DeleteOAuthAccount removes an OAuth account by its ID.
func (r *OAuthRepository) DeleteOAuthAccount(
	ctx context.Context,
	id pgtype.UUID,
) error {
	return r.queries.DeleteOAuthAccount(ctx, id)
}

// DeleteOAuthAccountByUserAndProvider removes a provider
// connection from a user.
func (r *OAuthRepository) DeleteOAuthAccountByUserAndProvider(
	ctx context.Context,
	userID pgtype.UUID,
	provider string,
) error {
	return r.queries.DeleteOAuthAccountByUserAndProvider(
		ctx,
		database.DeleteOAuthAccountByUserAndProviderParams{
			UserID:   userID,
			Provider: provider,
		},
	)
}
