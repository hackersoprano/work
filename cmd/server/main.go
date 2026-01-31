package main

import (
	"log"
	"work/api"
	"work/storages/postgres"
)

func main() {
	// Инициализация БД
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Настройка маршрутов
	router := api.SetupRoutes()

	// Запуск сервера
	log.Println("Starting server on :8080")
	if err := router.Start(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
