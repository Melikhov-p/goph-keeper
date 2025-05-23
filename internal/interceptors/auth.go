// Package interceptors пакет с перехватчиками для gRPC запросов.
package interceptors

import (
	"context"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/auth"
	contextkeys "github.com/Melikhov-p/goph-keeper/internal/context_keys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor перехватчик авторизации.
func AuthInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Пропускаем аутентификацию для публичных методов
		if info.FullMethod == "/gophkeeper.v1.UserService/Register" ||
			info.FullMethod == "/gophkeeper.v1.UserService/Login" {
			return handler(ctx, req)
		}

		// Извлекаем токен из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token not provided: %v")
		}

		token := authHeader[0]

		// Валидация токена
		userID, err := auth.GetUserIDbyToken(token, secretKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
		}

		// Добавляем userID в контекст
		newCtx := context.WithValue(ctx, contextkeys.UserID, userID)

		return handler(newCtx, req)
	}
}
