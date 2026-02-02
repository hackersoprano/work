package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"work/models"
	"work/services"

	"github.com/labstack/echo/v4"
)

func GetAll(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), GetTimeout)
	defer cancel()

	var allUsers []models.AllUser //создание пустого массива

	// Select автоматически сканирует результаты
	err := db.SelectContext(ctx, &allUsers, "SELECT id,login,role FROM users ORDER BY login") //запрос select
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Используем StatusOK для GET запросов
	return c.JSON(http.StatusOK, allUsers)
}

func CreateUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), PostTimeout)
	defer cancel()
	user := new(models.User) //создаем пустую структуру

	if err := c.Bind(user); err != nil { //заполняем из json запроса
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недопсутимое значение"})
	}
	// Проверяем обязательные поля
	if user.Login == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Логин и пароль обязательны",
		})
	}
	//хэшируем пароль
	user.Password = services.HashPassword(user.Password)

	// Named query с использованием структуры
	query := `INSERT INTO users (login, password, role) 
		          VALUES (:login, :password, :role) 
		          RETURNING id, login, password, role` //возвращение созданную запись

	// NamedQuery + StructScan для удобной работы
	rows, err := db.NamedQueryContext(ctx, query, user) //выполняем запрос и берем данные из user
	if err != nil {
		// Проверяем, если это ошибка уникальности
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Пользователь уже существует",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка базы данных"})
	}
	defer rows.Close() //обязательно закрываем строки результата, даже при ошибке

	// Сканируем возвращенные значения (включая сгенерированный ID)
	if rows.Next() { //переход на след строку
		err = rows.StructScan(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	user.Password = "" // очищаем значение для безопасности, перед выводом
	return c.JSON(http.StatusCreated, user)
}

func UpdateUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), PostTimeout)
	defer cancel()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка ID формата"})
	}

	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недопустимое значение"})
	}

	// Проверяем существование сотрудника
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE id = $1", id)
	if err != nil || count == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Данного пользователя не существует"})
	}

	user.ID = id
	//проверяем заполнение login
	var currentUser models.User
	err = db.GetContext(ctx, &currentUser, "SELECT id, login, role FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Данного пользователя не существует"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка базы данных"})
	}
	//
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
		user.Password = services.HashPassword(user.Password)
	}
	query := `UPDATE users 
              SET login = :login, password = :password, role = :role 
              WHERE id = :id`
	result, err := db.NamedExecContext(ctx, query, user)

	if err != nil {
		// Проверяем, возможно, проблема с уникальностью логина
		if strings.Contains(err.Error(), "unique") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Пользователь с таким логином уже существует",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка базы данных"})
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка базы данных"})
	}

	if rowsAffected == 0 {
		// На случай, если пользователь был удален параллельно
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}

	// Получаем обновленные данные
	err = db.GetContext(ctx, user, "SELECT id, login, password, role FROM users WHERE id = $1", id)
	if err != nil {
		// Даже если запись не найдена после обновления - это странно, но обрабатываем
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка базы данных"})
	}
	user.Password = "" //скрываем пароль(хэш)
	return c.JSON(http.StatusOK, user)
}
func DeleteUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), PostTimeout)
	defer cancel()
	id := c.Param("id")

	// Проверяем существование сотрудника
	var user models.User
	err := db.GetContext(ctx, &user, "SELECT id, login, password, role FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	//осторожность в удаление себя
	currentUserID := c.Get("user_id").(int)
	if user.ID == currentUserID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Нельзя удалить самого себя",
		})
	}

	// Удаляем сотрудника
	result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Проверяем количество удаленных строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Сотрудник успешно удален",
		"deleted_user": user,
	})
}
