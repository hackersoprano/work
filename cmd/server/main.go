package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"work/api"
	"work/services"
	"work/storages/postgres"
)

func main() {
	// Инициализация БД
	//db, err := postgres.NewConnection()
	//if err != nil {
	//	log.Fatal("Failed to connect to database:", err)
	//}
	//defer db.Close()

	storage, err := postgres.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer storage.Close()

	userService := services.NewUserService(storage)
	api.SetService(userService)
	server := api.New(userService)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting server")
		_ = server.Run(":8080")
	}()

	<-ctx.Done()

	log.Println("Shutting down server")
	if err = server.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}
