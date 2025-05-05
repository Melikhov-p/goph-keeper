// Package auth пакет с методами для генерации/разборки токенов.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims структура утверждений для JWT токена.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTToken строит JWT токен.
func BuildJWTToken(userID int, secretKey string, tokenLifeTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifeTime)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error signed string from []byte %w", err)
	}

	return tokenString, nil
}

// GetUserIDbyToken вытаскивает из токена ID пользователя.
func GetUserIDbyToken(tokenString string, secretKey string) (int, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected token singing method: %s", token.Method)
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return -1, errors.New("token expired")
		}
		return -1, fmt.Errorf("error parsing token with claims %w", err)
	}

	if !token.Valid {
		return -1, errors.New("invalid token")
	}

	return claims.UserID, nil
}
