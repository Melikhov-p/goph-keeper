package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
)

// ErrSecretNotFound секрет не найден.
var ErrSecretNotFound = errors.New("secret not found")

// Service структура сервиса.
type Service struct {
	repo Repository
	cfg  *config.Config
}

// NewService получение сервиса для секретов.
func NewService(r Repository, c *config.Config) *Service {
	return &Service{
		repo: r,
		cfg:  c,
	}
}

// CreateSecretPassword создать секретный пароль.
func (s *Service) CreateSecretPassword(
	ctx context.Context,
	u *user.User,
	secretName, username, password, url, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domani.Service.CreateSecretPassword"

	var (
		secret *Secret
		err    error
	)

	secret, err = NewPasswordSecret(u, secretName, username, password, url, notes, metaData, s.cfg.Security.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for password secret %w", op, err)
	}

	err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	return secret, nil
}

// CreateSecretCard создать секретные данные банковской карты.
func (s *Service) CreateSecretCard(
	ctx context.Context,
	u *user.User,
	secretName, number, owner, expireDate, cvv, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.CreateSecretCard"

	var (
		secret *Secret
		err    error
	)

	secret, err = NewCardSecret(u, secretName, number, owner, expireDate, cvv, notes, metaData, s.cfg.Security.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for card secret %w", op, err)
	}

	err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	return secret, nil
}

// CreateSecretFile создать новый секретный файл / двоичную информацию.
func (s *Service) CreateSecretFile(
	ctx context.Context,
	u *user.User,
	secretName, fileName, notes string,
	content, metaData []byte,
) (*Secret, error) {
	op := "domain.service.CreateSecretFile"

	var (
		secret *Secret
		err    error
	)

	secret, err = NewFileSecret(
		u,
		secretName,
		s.cfg.Database.ExternalStoragePath,
		fileName,
		content,
		notes,
		metaData,
		s.cfg.Security.MasterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for file secret %w", op, err)
	}

	err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	return secret, nil
}

// GetSecretsByName получить секреты по названию.
func (s *Service) GetSecretsByName(
	ctx context.Context,
	u *user.User,
	secretName string,
) ([]*Secret, error) {
	op := "domain.service.GetSecretByName"

	var (
		secrets []*Secret
		err     error
	)

	secrets, err = s.repo.GetSecretsByName(ctx, secretName, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to get secret by name %s with error %w", op, secretName, err)
	}

	if len(secrets) == 0 {
		return nil, ErrSecretNotFound
	}

	for _, secret := range secrets {
		secret.Data.setMasterKey(s.cfg.Security.MasterKey)
		err = secret.DecryptData()
		if err != nil {
			return nil, fmt.Errorf("%s: failed to decrypt data with error %w", op, err)
		}
	}

	return secrets, nil
}

func (s *Service) GetAllUserSecrets(ctx context.Context, u *user.User) ([]*Secret, error) {
	op := "domain.Secret.service.GetAllUserService"

	var (
		secrets []*Secret
		err     error
	)

	secrets, err = s.repo.GetAllUserSecrets(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get all secrets with error %w", op, err)
	}

	return secrets, nil
}
