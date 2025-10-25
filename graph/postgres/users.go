package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"example/hello/graph/model"

	"github.com/go-pg/pg/v10"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func (u *UsersRepo) CreateUser(ctx context.Context, input *model.NewUser) (*model.User, error) {
	dbUser := GraphQLUserToDBUser(input)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dbUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	dbUser.Password = string(hashedPassword)
	_, err = u.DB.Model(dbUser).Insert()
	if err != nil {
		return nil, err
	}
	return DBUserToGraphQLUser(dbUser), nil
}

func (u *UsersRepo) UpdateUser(ctx context.Context, id string, name *string, password *string, email *string) (*model.User, error) {
	userId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	dbUser := &DBUser{ID: userId}
	if err := u.DB.Model(dbUser).WherePK().Select(); err != nil {
		return nil, err
	}

	if name != nil {
		dbUser.Name = *name
	}
	if password != nil {
		hashedPw, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		dbUser.Password = string(hashedPw)
	}
	if email != nil {
		dbUser.Email = *email
	}

	if _, err := u.DB.Model(dbUser).WherePK().Update(); err != nil {
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
	if _, err := u.DB.Model(dbUser).WherePK().Delete(); err != nil {
		return "", err
	}
	return id, nil
}
func (u *UsersRepo) Login(ctx context.Context, email string, password string) (*model.AuthPayload, error) {
	// Debug: log receiver/DB pointers to help diagnose nil deref panics
	fmt.Printf("[DEBUG] UsersRepo.Login called - repo ptr=%p db ptr=%p email=%s\n", u, u.DB, email)

	// Guard against uninitialized repository/DB to avoid nil pointer dereference
	if u == nil || u.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dbUser := &DBUser{}
	err := u.DB.Model(dbUser).Where("email = ?", email).Select()
	if err != nil {
		// Treat no rows as invalid credentials rather than returning a DB error
		if err == pg.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, err
	}

	// Ensure a password hash exists before attempting comparison
	if dbUser == nil || dbUser.Password == "" {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	user := DBUserToGraphQLUser(dbUser)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET not set")
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenObj.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &model.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}
