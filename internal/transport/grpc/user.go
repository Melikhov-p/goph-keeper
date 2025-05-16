// Package grpc пакет обработчиков gRPC запросов.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/gen"
	"github.com/Melikhov-p/goph-keeper/internal/auth"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"github.com/Melikhov-p/goph-keeper/internal/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserService интерфейс сервиса пользователя.
type UserService interface {
	Register(ctx context.Context, login, password, pepper string) (*user.User, error)
	Login(ctx context.Context, login, password, pepper string) (*user.User, error)
	Update(ctx context.Context, u *user.User) error
}

// UserServer gRPC обработчик запросов для методов пользователя.
type UserServer struct {
	pb.UnimplementedUserServiceServer
	service UserService
	log     *zap.Logger
	cfg     *config.Config
}

// NewUserServer новый gRPC обработчик для пользователя.
func NewUserServer(us UserService, log *zap.Logger, cfg *config.Config) *UserServer {
	return &UserServer{
		service: us,
		log:     log,
		cfg:     cfg,
	}
}

// Register регистрация пользователя.
func (us *UserServer) Register(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	var (
		u   *user.User
		res pb.RegisterUserResponse
		err error
	)

	u, err = us.service.Register(ctx, in.GetLogin(), in.GetPassword(), us.cfg.Security.Pepper)
	if err != nil {
		us.log.Error("error register new user", zap.Error(err), zap.String("Login", in.GetLogin()))
		if errors.Is(err, user.ErrAlreadyExist) {
			err = status.Error(codes.AlreadyExists, "user already exist")
			return nil, fmt.Errorf("failed to register new user: %w", err)
		}
		err = status.Error(codes.Internal, "failed to register")
		return nil, fmt.Errorf("failed to register new user: %w", err)
	}

	err = addTokenToCtx(ctx, u.ID, us.cfg.Security.TokenKey, us.cfg.Security.TokenTTL)
	if err != nil {
		us.log.Error("error adding token to context for new user", zap.Error(err), zap.Int("UserID", u.ID))
	}

	userID32, err := util.SafeConvertToInt32(u.ID)
	if err != nil {
		us.log.Error("error convert userID to int32", zap.Error(err), zap.Int("UserID", u.ID))
		err = status.Error(codes.Internal, "failed to build response")
		return nil, fmt.Errorf("failed to write response: %w", err)
	}

	res.User = &pb.User{
		Login: u.Login,
		Id:    userID32,
	}

	return &res, nil
}

// Login аутентификация пользователя.
func (us *UserServer) Login(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	var (
		u   *user.User
		res pb.LoginUserResponse
		err error
	)

	u, err = us.service.Login(ctx, in.GetLogin(), in.GetPassword(), us.cfg.Security.Pepper)
	if err != nil {
		us.log.Error("error login user", zap.Error(err), zap.String("Login", in.GetLogin()))
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			err = status.Error(codes.InvalidArgument, "invalid credentials")
			return nil, fmt.Errorf("failed to login user: %w", err)
		case errors.Is(err, user.ErrNotFound):
			err = status.Error(codes.NotFound, "user not found")
			return nil, fmt.Errorf("failed to login user: %w", err)
		default:
			err = status.Error(codes.Internal, "failed to login")
			return nil, fmt.Errorf("failed to login user: %w", err)
		}
	}

	err = addTokenToCtx(ctx, u.ID, us.cfg.Security.TokenKey, us.cfg.Security.TokenTTL)
	if err != nil {
		us.log.Error("failed to add token for login user to context", zap.Error(err))
	}

	userID32, err := util.SafeConvertToInt32(u.ID)
	if err != nil {
		us.log.Error("error convert userID to int32", zap.Error(err), zap.Int("UserID", u.ID))
		err = status.Error(codes.Internal, "failed to build response")
		return nil, fmt.Errorf("failed to write response: %w", err)
	}

	res.User = &pb.User{
		Login: u.Login,
		Id:    userID32,
	}

	return &res, nil
}

func addTokenToCtx(ctx context.Context, userID int, tokenSecret string, tokenTTL time.Duration) error {
	op := "transport.gRPC.user.addTokenToCtx"

	var (
		token string
		err   error
	)

	token, err = auth.BuildJWTToken(userID, tokenSecret, tokenTTL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"authorization": token,
	}))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
