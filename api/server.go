package api

import (
	"context"

	"github.com/labstack/echo/v4"
)

type Server struct {
	e *echo.Echo

	user UserService
}

func New(service UserService) *Server {
	e := echo.New()

	s := &Server{
		e:    e,
		user: service,
	}

	s.SetupRoutes()

	return s
}

func (s *Server) Run(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
