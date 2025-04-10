package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// AuthInterceptor перехватчик авторизации.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// Пропускаем аутентификацию для публичных методов
		if info.FullMethod == "/gophkeeper.v1.UserService/Register" ||
			info.FullMethod == "/gophkeeper.v1.UserService/Login" {
			return handler(ctx, req)
		}

		return handler(ctx, req)
	}
}
