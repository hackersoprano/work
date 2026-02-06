package services

import (
	"context"
	"work/models"
)

type (
	Storage interface {
		GetUserByLogin(ctx context.Context, login string) (*models.User, error)
		GetAllUsers(ctx context.Context) ([]models.AllUser, error)
		CreateUser(ctx context.Context, user *models.User) error
		UpdateUser(ctx context.Context, id int, user *models.User) error
		DeleteUser(ctx context.Context, id int) error
	}
)
