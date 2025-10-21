package postgres

import (
	"context"

	"example/hello/graph/model"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type UsersRepo struct {
	DB *pg.DB
}

func (u *UsersRepo) GetUsers() ([]*model.User, error) {
	var userCollection []*DBUser
	err := u.DB.Model(&userCollection).Select()

	if err != nil {
		return nil, err
	}

	var users []*model.User
	for i := 0; i < len(userCollection); i++ {
		users = append(users, DBUserToGraphQLUser(userCollection[i]))
	}

	return users, nil
}

func (u *UsersRepo) GetUserByID(id string) (*model.User, error) {
	userId, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}
	var dbUser *DBUser = &DBUser{ID: userId}
	err = u.DB.Model(dbUser).WherePK().Select()
	if err != nil {
		return nil, err
	}
	return DBUserToGraphQLUser(dbUser), nil
}

func (u *UsersRepo) GetUserByEmail(email string) (*model.User, error) {
	var dbUser *DBUser = &DBUser{}
	err := u.DB.Model(dbUser).Where("email = ?", email).Select()
	if err != nil {
		return nil, err
	}
	return DBUserToGraphQLUser(dbUser), nil
}

func (u *UsersRepo) GetUserByName(name string) (*model.User, error) {
	var dbUser *DBUser = &DBUser{}

	err := u.DB.Model(dbUser).Where("name = ?", name).Select()
	if err != nil {
		return nil, err
	}
	return DBUserToGraphQLUser(dbUser), nil
}

func (u *UsersRepo) CreateUser(ctx context.Context, newUser *model.NewUser) (*model.User, error) {
	dbUser := GraphQLUserToDBUser(newUser)
	user := &model.User{
		ID:       dbUser.ID.String(),
		Name:     dbUser.Name,
		Email:    dbUser.Email,
		Password: dbUser.Password,
	}

	_, err := u.DB.Model(dbUser).Insert()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UsersRepo) UpdateUser(ctx context.Context, id string, name *string, password *string, email *string) (*model.User, error) {
	userId, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	dbUser := &DBUser{ID: userId}
	err = u.DB.Model(dbUser).WherePK().Select()

	if err != nil {
		return nil, err
	}
	if name != nil {
		dbUser.Name = *name
	}
	if password != nil {
		dbUser.Password = *password
	}
	if email != nil {
		dbUser.Email = *email
	}
	_, err = u.DB.Model(dbUser).WherePK().Update()
	if err != nil {
		return nil, err
	}
	return DBUserToGraphQLUser(dbUser), nil
}

func (u *UsersRepo) DeleteUser(ctx context.Context, id string) (string, error) {
	userId, err := uuid.Parse(id)

	if err != nil {
		return "", err
	}

	dbUser := &DBUser{ID: userId}
	_, err = u.DB.Model(dbUser).WherePK().Delete()
	if err != nil {
		return "", err
	}
	return id, nil
}
