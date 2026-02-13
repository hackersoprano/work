package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"work/api"
	"work/services"
	"work/storages/postgres"
)

const migrationsDir = "migrations"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

func main() {
	// восстановить миграцию
	migrator := postgres.MustGetNewMigrator(MigrationsFS, migrationsDir)
	// Инициализация БД

	storage, err := postgres.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer storage.Close()

	//приминение миграции
	err = migrator.ApplyMigrations(storage)
	if err != nil {
		panic(err)
	}
	log.Printf("Миграции применены!!")

	userService := services.NewUserService(storage)
	api.SetService(userService)
	server := api.New(userService)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting server")
		if err = server.Run(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down server")
	if err = server.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}
