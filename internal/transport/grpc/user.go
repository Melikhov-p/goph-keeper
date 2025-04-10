package grpc

import (
	"context"
	"errors"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/gen"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
			return nil, status.Error(codes.AlreadyExists, "user already exist")
		}
		return nil, status.Error(codes.Internal, "failed to register")
	}

	res.User = &pb.User{
		Login: u.Login,
		Id:    int32(u.ID),
	}

	return &res, nil
}

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

func (us *UserServer) Update(_ context.Context, _ *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "method is not implemented")
}
