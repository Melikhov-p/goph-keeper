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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
		u     *user.User
		token string
		res   pb.RegisterUserResponse
		err   error
	)

	u, err = us.service.Register(ctx, in.GetLogin(), in.GetPassword(), us.cfg.Security.Pepper)
	if err != nil {
		us.log.Error("error register new user", zap.Error(err), zap.String("Login", in.GetLogin()))
		if errors.Is(err, user.ErrAlreadyExist) {
			return nil, status.Error(codes.AlreadyExists, "user already exist")
		}
		return nil, status.Error(codes.Internal, "failed to register")
	}

	err = addTokenToCtx(ctx, u.ID, us.cfg.Security.TokenKey, us.cfg.Security.TokenTTL)
	if err != nil {
		us.log.Error("error adding token to context for new user", zap.Error(err), zap.Int("UserID", u.ID))
	}

	res.User = &pb.User{
		Login: u.Login,
		Id:    int32(u.ID),
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
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		case errors.Is(err, user.ErrNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "failed to login")
		}
	}

	res.User = &pb.User{
		Login: u.Login,
		Id:    int32(u.ID),
	}

	return &res, nil
}

// Update обновление пользователя.
func (us *UserServer) Update(_ context.Context, _ *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "method is not implemented")
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
