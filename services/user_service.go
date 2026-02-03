package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"work/models"

	"github.com/jmoiron/sqlx"
)

type UserServiceDb struct {
	db *sqlx.DB
}

func NewUserService(db *sqlx.DB) *UserServiceDb {
	return &UserServiceDb{db: db}
}

//метод авторизации

func (s *UserServiceDb) Authenticate(ctx context.Context, login, password string) (*models.User, error) {
	var user models.User

	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE login = $1", login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	// Проверяем пароль
	hashedPassword := HashPassword(password)
	if user.Password != hashedPassword {
		return nil, errors.New("неверный пароль")
	}
	return &user, nil
}

func (s *UserServiceDb) GetAllUsers(ctx context.Context) ([]models.AllUser, error) {
	var users []models.AllUser
	err := s.db.SelectContext(ctx, &users, "SELECT id, login, role FROM users ORDER BY login")
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserServiceDb) CreateUser(ctx context.Context, user *models.User) error {
	// Хэшируем пароль
	user.Password = HashPassword(user.Password)

	query := `INSERT INTO users (login, password, role) 
	          VALUES (:login, :password, :role) 
	          RETURNING id`

	rows, err := s.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return errors.New("пользователь с таким логином уже существует")
		}
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserServiceDb) UpdateUser(ctx context.Context, id int, user *models.User) error {
	user.ID = id
	//------------------------TODO---------------------------------переделать обновление пароля
	if user.Password != "" {
		user.Password = HashPassword(user.Password)
		query := `UPDATE users SET login = :login, password = :password, role = :role WHERE id = :id`
		_, err := s.db.NamedExecContext(ctx, query, user)
		return err
	}

	query := `UPDATE users SET login = :login, role = :role WHERE id = :id`
	_, err := s.db.NamedExecContext(ctx, query, user)
	return err
}

func (s *UserServiceDb) DeleteUser(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
