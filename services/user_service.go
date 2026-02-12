package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"work/models"
)

type UserServiceDb struct {
	db Storage
}

func NewUserService(db Storage) *UserServiceDb {
	return &UserServiceDb{db: db}
}

//метод авторизации

func (s *UserServiceDb) Authenticate(ctx context.Context, login, password string) (*models.User, error) {

	user, err := s.db.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("Пользователь не найден")
	}

	// Проверяем пароль
	hashedPassword := HashPassword(password)
	if user.Password != hashedPassword {
		return nil, errors.New("неверный пароль")
	}
	return user, nil
}

func (s *UserServiceDb) GetAllUsers(ctx context.Context) ([]models.AllUser, error) {
	return s.db.GetAllUsers(ctx)
}

func (s *UserServiceDb) CreateUser(ctx context.Context, user *models.User) error {
	tx, txCtx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // откат, если не сделан Commit
	// Хэшируем пароль
	user.Password = HashPassword(user.Password)
	if user.Role == "" {
		user.Role = "user"
	}
	err = s.db.CreateUser(txCtx, user)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return errors.New("пользователь уже существует")
		}
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *UserServiceDb) UpdateUser(ctx context.Context, user *models.User) error {
	tx, txCtx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	currentUser, err := s.db.GetUserById(txCtx, user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if user.Login == "" {
		user.Login = currentUser.Login
	}
	if user.Role == "" {
		user.Role = currentUser.Role
	}

	//проверка пароль изменен или нет.

	if user.Password == "" {
		user.Password = currentUser.Password
	} else {
		user.Password = HashPassword(user.Password)
	}
	err = s.db.UpdateUser(txCtx, user)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *UserServiceDb) DeleteUser(ctx context.Context, id int) error {
	tx, txCtx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = s.db.DeleteUser(txCtx, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
