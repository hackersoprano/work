package api

import (
	"fmt"
	"net/http"
	"strings"
	"work/models"
	"work/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization") //проверяем есть ли в заголовке запроса авторизация
		if authHeader == "" {                                 //если отсутствует требуем аавторизацию
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Требуется авторизация",
			})
		}
		//проверка формата токена
		parts := strings.Split(authHeader, " ")      //делим строку на 2 части по пробелу
		if len(parts) != 2 || parts[0] != "Bearer" { //первое слово должно быть барьер
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Неверный формат токена",
			})
		}

		tokenString := parts[1] //записываем токен в переменную

		// Проверяем токен
		token, err := jwt.ParseWithClaims(tokenString, &models.JwtUser{}, func(token *jwt.Token) (interface{}, error) { //
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return services.JwtSecret, nil
		})

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Неверный или истекший токен",
			})
		}
		//извлекаем данные о пользователе
		if claims, ok := token.Claims.(*models.JwtUser); ok && token.Valid {
			// Сохраняем данные пользователя в контекст
			c.Set("user_id", claims.UserID)
			c.Set("user_login", claims.Login)
			c.Set("user_role", claims.Role)
		} else {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Невалидный токен",
			})
		}

		return next(c) //если все ок, то пропускаем дальше
	}
}

// Middleware для проверки роли админа
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("user_role").(string) //получаем из "пакета" информацию о роле пользователя
		if role != "admin" {                //если роль не админ, запрещаем доступ
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Недостаточно прав. Требуется роль admin",
			})
		}
		return next(c) //если все ок, то пропускаем дальше
	}
}
