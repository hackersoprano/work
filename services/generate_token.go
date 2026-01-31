package services

import (
	"fmt"
	"os"
	"time"
	"work/models"

	"github.com/golang-jwt/jwt/v5"
)

// забираем jwt_secret из окружения
var JwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(userID int, login string, role string) (string, error) {
	if len(JwtSecret) == 0 {
		return "", fmt.Errorf("JWT_SECRET не установлен")
	}
	claims := &models.Jwt_user{ //формируем "пакет с данными"
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
	return token.SignedString(JwtSecret)                       //возврат токена или ошибки
}
