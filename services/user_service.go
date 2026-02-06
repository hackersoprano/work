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
	// Хэшируем пароль
	user.Password = HashPassword(user.Password)
	if user.Role == "" {
		user.Role = "user"
	}
	err := s.db.CreateUser(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return errors.New("пользователь уже существует")
		}
		return err
	}
	//defer rows.Close()
	//
	//if rows.Next() {
	//	err = rows.Scan(&user.ID)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (s *UserServiceDb) UpdateUser(ctx context.Context, id int, user *models.User) error {
	currentUser, err := s.db.GetUserByLogin(ctx, user.Login)
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
	return s.db.UpdateUser(ctx, id, user)
}

func (s *UserServiceDb) DeleteUser(ctx context.Context, id int) error {
	return s.db.DeleteUser(ctx, id)
}
