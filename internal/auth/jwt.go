package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

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

func GetUserIDbyToken(tokenString string, secretKey string) (int, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected token singing method: %s", token.Method)
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, fmt.Errorf("error parsing token with claims %w", err)
	}

	if !token.Valid {
		return -1, errors.New("invalid token")
	}

	return claims.UserID, nil
}
