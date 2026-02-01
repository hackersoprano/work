package api

import (
	"context"
	"net/http"
	"time"
	"work/models"
	"work/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

const (
	PostTimeout = 5 * time.Second  //время на отправку
	GetTimeout  = 10 * time.Second //время на получение информации
)

var db *sqlx.DB

func SetDB(database *sqlx.DB) {
	db = database
}
func Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil { //получение и преобразование из json в удобный для go структуру
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат данных",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), PostTimeout)
	defer cancel()

	// Ищем пользователя в базе
	var user models.User
	err := db.GetContext(ctx, &user,
		"SELECT * FROM users WHERE login = $1",
		req.Login)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Неверный логин или пароль ",
		})
	}
	// Проверяем пароль
	hashedPassword := services.HashPassword(req.Password)
	if user.Password != hashedPassword {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Неверный логин или пароль",
		})
	}

	// Генерируем JWT токен
	token, err := services.GenerateToken(user.ID, user.Login, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании токена",
		})
	}

	// Очищаем пароль перед отправкой
	user.Password = ""

	return c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}
