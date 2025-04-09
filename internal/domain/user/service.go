package user

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var (
	ErrUserAlreadyExist   = errors.New("user already exist")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service сервисный слой пользователя.
type Service struct {
	repo Repository
	log  *zap.Logger
}

// NewService возвращает указатель на сервис для пользователя.
func NewService(r Repository, l *zap.Logger) *Service {
	return &Service{
		repo: r,
		log:  l,
	}
}

func (s *Service) Register(ctx context.Context, login, password, pepper string) (*User, error) {
	op := "domain.User.Register"

	var (
		user *User
		err  error
	)

	user, err = s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check existing of user %w", op, err)
	}
	if user != nil {
		return nil, ErrUserAlreadyExist
	}

	user, err = NewUser(login, password, pepper)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get new domain model %w", op, err)
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create new user in repo %w", op, err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, login, password, pepper string) (*User, error) {
	var (
		user *User
		err  error
	)

	user, err = s.repo.GetByLogin(ctx, login)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.VerifyUserPassword(password, pepper) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
