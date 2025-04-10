package grpc

import (
	"context"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/proto"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, login, password, pepper string) (*user.User, error)
	Login(ctx context.Context, login, password, pepper string) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
}

type UserServer struct {
	pb.UnimplementedUserServiceServer
	service UserService
	log     *zap.Logger
	cfg     *config.Config
}

func NewUserServer(us UserService, log *zap.Logger, cfg *config.Config) *UserServer {
	return &UserServer{
		service: us,
		log:     log,
		cfg:     cfg,
	}
}
