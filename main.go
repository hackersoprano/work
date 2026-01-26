package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	//"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // PostgreSQL драйвер
)

// Глобальная переменная для подключения к БД
var db *sqlx.DB
var jwtSecret []byte

// Определение структуры для сотрудника
// sqlx теги для работы с базой данных
type User struct {
	ID       int    `json:"id" db:"id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
}

type jwt_user struct { //структура jwt токена
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
type LoginRequest struct { //структура авторизации
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct { //структура для вывода информации после авторизации
	Token string `json:"token"`
	User  User   `json:"user"`
}

// хэширование пароля
func hashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil)) //возврат значения и преобразование из бит в 16 с.ч.
}

// Генерация JWT токена
func generateToken(userID int, login string, role string) (string, error) {
	claims := &jwt_user{ //формируем "пакет с данными"
		UserID: userID,
		Login:  login,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), //срок действия
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     //когда(а именно сейчас)
			Subject:   login,                                              //в поле subject помещается Login = кому принадлежит
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //шифрование токена методом hs256
	return token.SignedString(jwtSecret)                       //возврат токена или ошибки
}

// Middleware для проверки JWT токена(авторизация с проверкой токена)
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
		token, err := jwt.ParseWithClaims(tokenString, &jwt_user{}, func(token *jwt.Token) (interface{}, error) { //
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Неверный или истекший токен",
			})
		}
		//извлекаем данные о пользователе
		if claims, ok := token.Claims.(*jwt_user); ok && token.Valid {
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

//----------------------------------------------------------------------------------------------------------------------

func main() {

	var err error
	//забираем jwt_secret из окружения
	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		log.Println("Внимание: JWT_SECRET не задан")
	}
	jwtSecret = []byte(jwtSecretStr)

	// Подключение к базе данных PostgreSQL
	// изменение для dockerfile...
	//используем переменную окружения
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

	e := echo.New() //создание веб-сервера на фреймворке echo
	//-----------------Публичные ссылки-------------------------------
	// Обработчик GET запроса для получения всех сотрудников
	e.GET("/api/v1/users", func(c echo.Context) error {
		var users []User //создание пустого массива

		// Select автоматически сканирует результаты
		err := db.Select(&users, "SELECT login FROM users ORDER BY login") //запрос select
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Используем StatusOK для GET запросов
		return c.JSON(http.StatusOK, users)
	})

	//----------------------Авторизация-----------------------------------------_________________
	e.POST("/api/v1/login", func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil { //получение и преобразование из json в удобный для go структуру
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Неверный формат данных",
			})
		}
		// Ищем пользователя в базе
		var user User
		err := db.Get(&user,
			"SELECT * FROM users WHERE login = $1",
			req.Login)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Неверный логин или пароль ",
			})
		}
		// Проверяем пароль
		hashedPassword := hashPassword(req.Password)
		if user.Password != hashedPassword {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Неверный логин или пароль",
			})
		}

		// Генерируем JWT токен
		token, err := generateToken(user.ID, user.Login, user.Role)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при создании токена",
			})
		}

		// Очищаем пароль перед отправкой
		user.Password = ""

		return c.JSON(http.StatusOK, AuthResponse{
			Token: token,
			User:  user,
		})
	})

	//-------------------Приватные ссылки-----------------------------------
	//группа для авторизованных, в задании их нет---------------
	//authGroup := e.Group("/api/v1/")
	//authGroup.Use(AuthMiddleware)

	//группа для админов-----------------------------------
	adminGroup := e.Group("/api/v1/admin")
	adminGroup.Use(AuthMiddleware)
	adminGroup.Use(AdminMiddleware)

	// Обработчик POST запроса для создания нового сотрудника
	adminGroup.POST("/users", func(c echo.Context) error {
		user := new(User) //создаем пустую структуру

		if err := c.Bind(user); err != nil { //заполняем из json запроса
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недопсутимое значение"})
		}
		// Проверяем обязательные поля
		if user.Login == "" || user.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Логин и пароль обязательны",
			})
		}

		// Проверяем, существует ли пользователь!
		var exists bool                                                                           //true and false с этим помогает exist postgresql
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)", user.Login) //exists возращается true если произошло первое совпаеднеи
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка базы данных",
			})
		}

		if exists {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Пользователь уже существует",
			})
		}

		//если все ок -> хэшируем пароль
		user.Password = hashPassword(user.Password)

		// Named query с использованием структуры
		query := `INSERT INTO users (login, password, role) 
		          VALUES (:login, :password, :role) 
		          RETURNING id, login, password, role` //возвращение созданную запись

		// NamedQuery + StructScan для удобной работы
		rows, err := db.NamedQuery(query, user) //выполняем запрос и берем данные из user
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close() //обязательно закрываем строки результата, даже при ошибке

		// Сканируем возвращенные значения (включая сгенерированный ID)
		if rows.Next() { //переход на след строку
			err = rows.StructScan(user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		user.Password = "" // очищаем значение для безопасности, перед выводом
		return c.JSON(http.StatusCreated, user)
	})

	// Обработчик PUT запроса для обновления сотрудника
	adminGroup.PUT("/users/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка ID формата"})
		}

		user := new(User)
		if err := c.Bind(user); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Недопустимое значение"})
		}

		// Проверяем существование сотрудника
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1", id)
		if err != nil || count == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Данного пользователя не существует"})
		}

		user.ID = id
		//проверяем заполнение login
		var currentUser User
		err = db.Get(&currentUser, "SELECT id, login, role FROM users WHERE id = $1", id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Данного пользователя не существует"})
		}
		//
		if user.Login == "" {
			user.Login = currentUser.Login
		}
		if user.Role == "" {
			user.Role = currentUser.Role
		}

		//проверка пароль изменен или нет.
		if user.Password != "" {
			user.Password = hashPassword(user.Password)
			query := `UPDATE users 
		          SET login = :login, password = :password, role = :role 
		          WHERE id = :id`
			result, err := db.NamedExec(query, user)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			// Проверяем количество обновленных строк
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}

			if rowsAffected == 0 {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
			}
		} else {
			query := `UPDATE users 
		          SET login = :login, role = :role 
		          WHERE id = :id`
			result, err := db.NamedExec(query, user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			// Проверяем количество обновленных строк
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}

			if rowsAffected == 0 {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
			}
		}

		// Получаем обновленные данные
		err = db.Get(user, "SELECT id, login, password, role FROM users WHERE id = $1", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		user.Password = "" //скрываем пароль(хэш)
		return c.JSON(http.StatusOK, user)
	})

	// Обработчик DELETE запроса для удаления сотрудника
	adminGroup.DELETE("/users/:id", func(c echo.Context) error {
		id := c.Param("id")

		// Проверяем существование сотрудника
		var user User
		err := db.Get(&user, "SELECT id, login, password, role FROM users WHERE id = $1", id)
		if err != nil {
			if err.Error() == "sql: нет строк в результате запрос" {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		//осторожность в удаление себя
		currentUserID := c.Get("user_id").(int)
		if user.ID == currentUserID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Нельзя удалить самого себя",
			})
		}

		// Удаляем сотрудника
		result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Проверяем количество удаленных строк
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":      "Сотрудник успешно удален",
			"deleted_user": user,
		})
	})

	// Запуск сервера
	fmt.Println("Server starting on port 8080...")
	e.Logger.Fatal(e.Start(":8080"))
}
