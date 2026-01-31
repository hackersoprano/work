package models

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int    `json:"id" db:"id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
}

type Jwt_user struct { //структура jwt токена
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
