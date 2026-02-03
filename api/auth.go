package api

import (
	"context"
	"net/http"
	"time"
	"work/models"
	"work/services"

	"github.com/labstack/echo/v4"
)

const (
	PostTimeout = 5 * time.Second  //время на отправку
	GetTimeout  = 10 * time.Second //время на получение информации
)

var userService services.UserService

func SetService(userSvc services.UserService) {
	userService = userSvc
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

	user, err := userService.Authenticate(ctx, req.Login, req.Password)
	if err != nil {
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
		User:  *user,
	})
}
