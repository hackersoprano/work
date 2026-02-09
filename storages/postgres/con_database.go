package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"work/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB
var err error

type Storage struct {
	db *sqlx.DB
}

func NewConnection() (*Storage, error) {
	dbConnStr := os.Getenv("DATABASE_URL") //проверка переменной в docker
	if dbConnStr == "" {                   //если пустой, то вручную задаем
		dbConnStr = "postgresql://postgres:postgres@db:5432/workspace?sslmode=disable"
	}
	db, err = sqlx.Open("postgres", dbConnStr) //используем библиотеку postgres
	if err != nil {
		log.Fatal(err)
	}

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB Connected...")
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE login = $1", login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}
func (s *Storage) GetUserById(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users where id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Пользователь не найден")
		}
		return nil, err
	}
	return &user, nil
}
func (s *Storage) GetAllUsers(ctx context.Context) ([]models.AllUser, error) {
	var users []models.AllUser
	err := s.db.SelectContext(ctx, &users, "SELECT id, login, role FROM users ORDER BY login")
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (s *Storage) CreateUser(ctx context.Context, user *models.User) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `INSERT INTO users (login, password, role) 
	          VALUES (:login, :password, :role) 
	          RETURNING id`

	rows, err := tx.NamedQuery(query, user)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return errors.New("пользователь уже существует")
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
	return tx.Commit()
}
func (s *Storage) UpdateUser(ctx context.Context, user *models.User) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `UPDATE users 
              SET login = :login, password = :password, role = :role 
              WHERE id = :id`
	result, err := tx.NamedExecContext(ctx, query, user)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("ошибка обновления данных")
	}

	return tx.Commit()
}
func (s *Storage) DeleteUser(ctx context.Context, id int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("пользователь не найден")
	}
	return tx.Commit()
}
