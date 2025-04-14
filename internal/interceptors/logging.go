package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// LogInterceptor перехватчик логирования запросов.
func LogInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		startTime := time.Now()

		res, err := handler(ctx, req)

		duration := time.Since(startTime)

		log.Debug("",
			zap.String("method", info.FullMethod),
			zap.Any("response", res),
			zap.Duration("duration", duration),
			zap.Error(err))

		return res, err
	}
}
