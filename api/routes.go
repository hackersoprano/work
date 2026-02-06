package api

func (s *Server) SetupRoutes() {
	// Публичные маршруты
	s.e.GET("/api/v1/users", GetAll)
	s.e.POST("/api/v1/login", Login)

	// Защищенные маршруты (группы)
	adminGroup := s.e.Group("/api/v1/admin")
	adminGroup.Use(AuthMiddleware)
	adminGroup.Use(AdminMiddleware)

	adminGroup.POST("/users", CreateUser)
	adminGroup.PUT("/users/:id", UpdateUser)
	adminGroup.DELETE("/users/:id", DeleteUser)
}
