// Package app пакет с описанием приложения
package app

import (
	"context"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/Melikhov-p/goph-keeper/internal/logger"
	"github.com/Melikhov-p/goph-keeper/internal/repository/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// App структура приложения.
type App struct {
	UserRepository user.Repository

	UserService *user.Service

	GRPCServer *grpc.Server

	Log *zap.Logger
}

// New создание нового приложения.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	op := "app.New"
	app := App{}

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: error getting connection to db %w", op, err)
	}

	app.Log, err = logger.BuildLogger(cfg.Logging.Level)
	if err != nil {
		return nil, fmt.Errorf("%s: error getting logger %w", op, err)
	}

	app.UserRepository = postgres.NewUserRepository(db)
	app.UserService = user.NewService(app.UserRepository, app.Log)

	return &app, nil
}
