package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
)

var ErrSecretNotFound = errors.New("secret not found")

type Service struct {
	repo Repository
	cfg  *config.Config
}

func NewService(r Repository, c *config.Config) *Service {
	return &Service{
		repo: r,
		cfg:  c,
	}
}

func (s *Service) CreateSecretPassword(
	ctx context.Context,
	u user.User,
	secretName, username, password, url, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domani.Service.CreateSecretPassword"

	var (
		secret      *Secret
		newSecretID int
		err         error
	)

	secret, err = NewPasswordSecret(u, secretName, username, password, url, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for password secret %w", op, err)
	}

	newSecretID, err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	secret.SetID(newSecretID)

	return secret, nil
}

func (s *Service) CreateSecretCard(
	ctx context.Context,
	u user.User,
	secretName, number, owner, expireDate, cvv, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.CreateSecretCard"

	var (
		secret      *Secret
		newSecretID int
		err         error
	)

	secret, err = NewCardSecret(u, secretName, number, owner, expireDate, cvv, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for card secret %w", op, err)
	}

	newSecretID, err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	secret.SetID(newSecretID)

	return secret, nil
}

func (s *Service) CreateSecretFile(
	ctx context.Context,
	u user.User,
	secretName, fileName, content, notes string,
	metaData []byte,
) (*Secret, error) {
	op := "domain.service.CreateSecretFile"

	var (
		secret      *Secret
		newSecretID int
		err         error
	)

	secret, err = NewFileSecret(u, secretName, s.cfg.Database.ExternalStoragePath, fileName, content, notes, metaData)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model for file secret %w", op, err)
	}

	newSecretID, err = s.repo.SaveSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save secret on storage with error %w", op, err)
	}

	secret.SetID(newSecretID)

	return secret, nil
}

func (s *Service) GetSecretsByName(
	ctx context.Context,
	u user.User,
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

	return secrets, nil
}
