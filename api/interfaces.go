package api

import (
	"context"
	"work/models"
)

type (
	UserService interface {
		Authenticate(ctx context.Context, login, password string) (*models.User, error)
		GetAllUsers(ctx context.Context) ([]models.AllUser, error)
		CreateUser(ctx context.Context, user *models.User) error
		UpdateUser(ctx context.Context, id int, user *models.User) error
		DeleteUser(ctx context.Context, id int) error
	}
)
