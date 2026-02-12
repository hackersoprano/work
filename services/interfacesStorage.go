package services

import (
	"context"
	"database/sql"
	"work/models"
)

// Transaction определяет методы для управления транзакцией.
type Transaction interface {
	Commit() error
	Rollback() error
}
type (
	Storage interface {
		GetUserByLogin(ctx context.Context, login string) (*models.User, error)
		GetUserById(ctx context.Context, id int) (*models.User, error)
		GetAllUsers(ctx context.Context) ([]models.AllUser, error)
		CreateUser(ctx context.Context, user *models.User) error
		UpdateUser(ctx context.Context, user *models.User) error
		DeleteUser(ctx context.Context, id int) error
		BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, context.Context, error)
	}
)
