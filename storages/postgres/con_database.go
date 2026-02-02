package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var db *sqlx.DB
var err error

func NewConnection() (*sqlx.DB, error) {
	dbConnStr := os.Getenv("DATABASE_URL") //проверка переменной в docker
	if dbConnStr == "" {                   //если пустой, то вручную задаем
		dbConnStr = "postgresql://postgres:postgres@db:5432/workspace?sslmode=disable"
	}
	db, err = sqlx.Open("postgres", dbConnStr) //используем библиотеку postgres
	if err != nil {
		log.Fatal(err)
	}

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB Connected...")
	return db, nil
}
