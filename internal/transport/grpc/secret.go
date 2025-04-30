package grpc

import (
	"context"
	"fmt"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/gen"
	"github.com/Melikhov-p/goph-keeper/internal/config"
	contextkeys "github.com/Melikhov-p/goph-keeper/internal/context_keys"
	"github.com/Melikhov-p/goph-keeper/internal/domain/secret"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	secretTypePassword = iota
	secretTypeCard
	secretTypeBinary
)

var invalidSecretTypeErr = status.Error(codes.InvalidArgument, "invalid secret type")

// SecretCreator методы создания новых секретов.
type SecretCreator interface {
	CreateSecretPassword(
		ctx context.Context,
		u *user.User,
		secretName, username, password, url, notes string,
		metaData []byte,
	) (*secret.Secret, error)
	CreateSecretCard(
		ctx context.Context,
		u *user.User,
		secretName, number, owner, expireDate, cvv, notes string,
		metaData []byte,
	) (*secret.Secret, error)
	CreateSecretFile(
		ctx context.Context,
		u *user.User,
		secretName, fileName, notes string,
		content, metaData []byte,
	) (*secret.Secret, error)
}

// SecretProvider методы получения информации о секретах.
type SecretProvider interface {
	GetSecretsByName(
		ctx context.Context,
		u user.User,
		secretName string,
	) ([]*secret.Secret, error)
}

type UserProvider interface {
	GetUserByID(ctx context.Context, userID int) (*user.User, error)
}

// SecretServer обработчик запросов по секретам.
type SecretServer struct {
	pb.UnimplementedSecretServiceServer
	secretCreator  SecretCreator
	secretProvider SecretProvider
	userProvider   UserProvider
	cfg            *config.Config
	log            *zap.Logger
}

// NewSecretServer получение обработчика запросов для секретов.
func NewSecretServer(
	sC SecretCreator,
	sP SecretProvider,
	uP UserProvider,
	c *config.Config,
	l *zap.Logger,
) *SecretServer {
	return &SecretServer{
		secretCreator:  sC,
		secretProvider: sP,
		userProvider:   uP,
		cfg:            c,
		log:            l,
	}
}

// CreateSecret создание нового секрета.
func (ss *SecretServer) CreateSecret(ctx context.Context, in *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
	op := "transport.GRPC.Secret.Create"

	var (
		s         *secret.Secret
		u         *user.User
		res       pb.CreateSecretResponse
		logErrMsg string
		err       error
	)

	defer func() {
		if err != nil {
			ss.log.Error("error in CreateSecret", zap.Error(err), zap.String("note", logErrMsg))
		}
	}()

	userID, ok := ctx.Value(contextkeys.UserID).(int)
	if !ok {
		logErrMsg = "error getting userID from context with error"
		return nil, status.Error(codes.Unauthenticated, "user ID not found in auth token.")
	}
	u, err = ss.userProvider.GetUserByID(ctx, userID)
	if err != nil {
		logErrMsg = fmt.Sprintf("error getting user by id %d", userID)
		return nil, status.Error(codes.Unauthenticated, "user with provided ID not found")
	}

	switch in.Type {
	case secretTypePassword:
		data := in.GetPasswordData()
		s, err = ss.secretCreator.CreateSecretPassword(
			ctx, u, in.Name, data.Username, data.Password, data.Url, *data.Notes, data.MetaData,
		)
		if err != nil {
			logErrMsg = "error creating secret password"
			return nil, status.Error(codes.Internal, "failed to create new secret password.")
		}

		res.Id = int64(s.ID)
	case secretTypeCard:
		data := in.GetCardData()
		s, err = ss.secretCreator.CreateSecretCard(
			ctx, u, in.Name, data.Number, data.Owner, data.ExpireDate, data.CVV, *data.Notes, data.MetaData,
		)
		if err != nil {
			logErrMsg = "error creating secret card"
			return nil, status.Error(codes.Internal, "failed to create new card secret.")
		}

		res.Id = int64(s.ID)
	case secretTypeBinary:
		data := in.GetBinaryData()
		s, err = ss.secretCreator.CreateSecretFile(
			ctx, u, in.Name, data.Filename, *data.Notes, data.Content, data.MetaData,
		)
		if err != nil {
			logErrMsg = "error creating secret file"
			return nil, status.Error(codes.Internal, "failed to create new file secret.")
		}

		res.Id = int64(s.ID)
	default:
		logErrMsg = "invalid secret type"
		return nil, fmt.Errorf("%s: bad request: %w", op, err)
	}

	return &res, nil
}
