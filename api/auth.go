package api

import (
	"net/http"
	"work/models"
	"work/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
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
	// Ищем пользователя в базе
	var user models.User
	err := db.Get(&user,
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
