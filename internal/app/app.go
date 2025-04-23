// Package app пакет с описанием приложения
package app

import (
	"fmt"
	"net"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/gen"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/secret"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/Melikhov-p/goph-keeper/internal/interceptors"
	"github.com/Melikhov-p/goph-keeper/internal/logger"
	"github.com/Melikhov-p/goph-keeper/internal/repository/postgres"
	grpc2 "github.com/Melikhov-p/goph-keeper/internal/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// App структура приложения.
type App struct {
	Cfg *config.Config

	UserRepository user.Repository
	UserService    *user.Service

	SecretRepository secret.Repository
	SecretService    *secret.Service

	GRPCServer *grpc.Server

	Log *zap.Logger
}

// New создание нового приложения.
func New(cfg *config.Config) (*App, error) {
	op := "app.New"
	app := App{}
	app.Cfg = cfg

	db, err := postgres.NewConnection(app.Cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: error getting connection to db %w", op, err)
	}

	app.Log, err = logger.BuildLogger(app.Cfg.Logging.Level)
	if err != nil {
		return nil, fmt.Errorf("%s: error getting logger %w", op, err)
	}

	app.UserRepository = postgres.NewUserRepository(db)
	app.UserService = user.NewService(app.UserRepository)

	app.SecretRepository = postgres.NewSecretRepository(db)
	app.SecretService = secret.NewService(app.SecretRepository, app.Cfg)

	// Создание gRPC-сервера
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LogInterceptor(app.Log),
			interceptors.AuthInterceptor(cfg.Security.TokenKey),
		),
	)

	userServer := grpc2.NewUserServer(app.UserService, app.Log, app.Cfg)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	app.GRPCServer = grpcServer

	return &app, nil
}

// RunGRPC запуск gRPC сервера.
func (a *App) RunGRPC() error {
	op := "app.RunGRPC"

	listen, err := net.Listen("tcp", a.Cfg.RPC.Address)
	if err != nil {
		return fmt.Errorf("%s: failed to get listen for gRPC server %w", op, err)
	}

	if err = a.GRPCServer.Serve(listen); err != nil {
		return fmt.Errorf("%s: error serving gRPC server %w", op, err)
	}

	return nil
}

// StopGRPC graceful shutdown gRPC сервера.
func (a *App) StopGRPC() {
	a.GRPCServer.GracefulStop()
}
