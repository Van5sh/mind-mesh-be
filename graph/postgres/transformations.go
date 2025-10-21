package postgres

import (
	"example/hello/graph/model"

	"github.com/google/uuid"
)

func GraphQLUserToDBUser(NewUser *model.NewUser) *DBUser {
	return &DBUser{
		ID:       uuid.New(),
		Name:     NewUser.Name,
		Email:    NewUser.Email,
		Password: NewUser.Password,
	}
}

func DBUserToGraphQLUser(DBUser *DBUser) *model.User {
	return &model.User{
		ID:       DBUser.ID.String(),
		Name:     DBUser.Name,
		Email:    DBUser.Email,
		Password: DBUser.Password,
	}
}
