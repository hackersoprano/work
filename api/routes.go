package api

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes() *echo.Echo {
	e := echo.New()

	// Публичные маршруты
	e.GET("/api/v1/users", GetAll)
	e.POST("/api/v1/login", Login)

	// Защищенные маршруты (группы)
	adminGroup := e.Group("/api/v1/admin")
	adminGroup.Use(AuthMiddleware)
	adminGroup.Use(AdminMiddleware)

	adminGroup.POST("/users", CreateUser)
	adminGroup.PUT("/users/:id", UpdateUser)
	adminGroup.DELETE("/users/:id", DeleteUser)

	return e
}
