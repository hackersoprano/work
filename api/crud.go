package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"work/models"

	"github.com/labstack/echo/v4"
)

func GetAll(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), GetTimeout)
	defer cancel()
	allUsers, err := userService.GetAllUsers(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, allUsers)
}

func CreateUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), GetTimeout)
	defer cancel()
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{"error": "Недопустимое значение"})
	}

	if user.Login == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Логин и пароль обязательны",
		})
	}

	// Используем интерфейс UserService
	err := userService.CreateUser(ctx, user)
	if err != nil {
		if err.Error() == "пользователь с таким логином уже существует" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Пользователь уже существует",
			})
		}
		return c.JSON(http.StatusInternalServerError,
			map[string]string{"error": err.Error()})
	}
	user.Password = "" // очищаем значение для безопасности, перед выводом
	return c.JSON(http.StatusCreated, user)
}

func UpdateUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), GetTimeout)
	defer cancel()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{"error": "Ошибка ID формата"})
	}

	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{"error": "Недопустимое значение"})
	}
	user.ID = id
	err = userService.UpdateUser(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			map[string]string{"error": err.Error()})
	}
	user.Password = "" //скрываем пароль(хэш)
	return c.JSON(http.StatusOK, user)
}
func DeleteUser(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), GetTimeout)
	defer cancel()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{"error": "Ошибка ID формата"})
	}
	currentUserID := c.Get("user_id").(int)
	if id == currentUserID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Нельзя удалить самого себя",
		})
	}
	// Используем интерфейс UserService
	err = userService.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Пользователь удален",
	})
}
