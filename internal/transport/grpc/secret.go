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

// SecretService методы создания новых секретов.
type SecretService interface {
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
	GetSecretsByName(
		ctx context.Context,
		u *user.User,
		secretName string,
	) ([]*secret.Secret, error)
}

type UserProvider interface {
	GetUserByID(ctx context.Context, userID int) (*user.User, error)
}

// SecretServer обработчик запросов по секретам.
type SecretServer struct {
	pb.UnimplementedSecretServiceServer
	secretService SecretService
	userProvider  UserProvider
	cfg           *config.Config
	log           *zap.Logger
}

// NewSecretServer получение обработчика запросов для секретов.
func NewSecretServer(
	sC SecretService,
	uP UserProvider,
	c *config.Config,
	l *zap.Logger,
) *SecretServer {
	return &SecretServer{
		secretService: sC,
		userProvider:  uP,
		cfg:           c,
		log:           l,
	}
}

// CreateSecret создание нового секрета.
func (ss *SecretServer) CreateSecret(ctx context.Context, in *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
	op := "transport.GRPC.Secret.Create"

	var (
		s   *secret.Secret
		u   *user.User
		res pb.CreateSecretResponse
		err error
	)

	userID, ok := ctx.Value(contextkeys.UserID).(int)
	if !ok {
		ss.log.Error("error getting userID from context with error")
		return nil, status.Error(codes.Unauthenticated, "user ID not found in auth token.")
	}
	u, err = ss.userProvider.GetUserByID(ctx, userID)
	if err != nil {
		ss.log.Error("error getting user by id", zap.Int("ID", userID), zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user with provided ID not found")
	}

	switch in.Type {
	case secretTypePassword:
		data := in.GetPasswordData()
		s, err = ss.secretService.CreateSecretPassword(
			ctx, u, in.Name, data.Username, data.Password, data.Url, *data.Notes, data.MetaData,
		)
		if err != nil {
			ss.log.Error("error creating secret password", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create new secret password.")
		}

		res.Id = int64(s.ID)
	case secretTypeCard:
		data := in.GetCardData()
		s, err = ss.secretService.CreateSecretCard(
			ctx, u, in.Name, data.Number, data.Owner, data.ExpireDate, data.CVV, *data.Notes, data.MetaData,
		)
		if err != nil {
			ss.log.Error("error creating secret card", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create new card secret.")
		}

		res.Id = int64(s.ID)
	case secretTypeBinary:
		data := in.GetBinaryData()
		s, err = ss.secretService.CreateSecretFile(
			ctx, u, in.Name, data.Filename, *data.Notes, data.Content, data.MetaData,
		)
		if err != nil {
			ss.log.Error("error creating secret file", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create new file secret.")
		}

		res.Id = int64(s.ID)
	default:
		ss.log.Warn("invalid secret type", zap.Any("type", in.Type))
		return nil, fmt.Errorf("%s: bad request: %w", op, err)
	}

	return &res, nil
}

func (ss *SecretServer) GetSecret(ctx context.Context, in *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	var (
		res        pb.GetSecretResponse
		s          []*secret.Secret
		u          *user.User
		secretName string
		err        error
	)

	userID, ok := ctx.Value(contextkeys.UserID).(int)
	if !ok {
		ss.log.Error("error getting userID from context with error")
		return nil, status.Error(codes.Unauthenticated, "user ID not found in auth token.")
	}
	u, err = ss.userProvider.GetUserByID(ctx, userID)
	if err != nil {
		ss.log.Error("error getting user by id", zap.Int("ID", userID), zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user with provided ID not found")
	}

	if secretName = in.GetName(); secretName != "" {
		s, err = ss.secretService.GetSecretsByName(ctx, u, secretName)
		if err != nil {
			ss.log.Debug("secrets not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "secrets not found")
		}
	}

	ss.log.Debug("found secrets", zap.Any("secrets", s))

	if len(s) == 0 {
		ss.log.Debug("len secrets is 0")
		return nil, status.Error(codes.NotFound, "secrets not found")
	}

	for _, sec := range s {
		foundResSecret := pb.GetSecret{}
		foundResSecret.Name = sec.Name
		switch sec.Type {
		case secret.TypePassword:
			data, _ := sec.Data.(*secret.PasswordData)
			foundResSecret.Type = secretTypePassword
			foundResSecret.Data = &pb.GetSecret_PasswordData{
				PasswordData: &pb.PasswordData{
					Username: data.Username,
					Password: data.Pass,
					Url:      data.URL,
					Notes:    &data.Notes,
					MetaData: data.MetaData,
				},
			}
		case secret.TypeCard:
			data, _ := sec.Data.(*secret.CardData)
			foundResSecret.Type = secretTypeCard
			foundResSecret.Data = &pb.GetSecret_CardData{
				CardData: &pb.CardData{
					Owner:      data.Owner,
					CVV:        data.CVV,
					ExpireDate: data.ExpireDate,
					Number:     data.Number,
					MetaData:   data.MetaData,
					Notes:      &data.Notes,
				},
			}
		case secret.TypeBinary:
			data, _ := sec.Data.(*secret.FileData)
			foundResSecret.Type = secretTypeBinary
			foundResSecret.Data = &pb.GetSecret_BinaryData{
				BinaryData: &pb.BinaryData{
					Filename: data.Name,
					Content:  data.Content,
					MetaData: data.MetaData,
					Notes:    &data.Notes,
				},
			}
		}

		res.Secrets = append(res.Secrets, &foundResSecret)
	}

	return &res, nil
}
