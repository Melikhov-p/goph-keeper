package user

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var (
	// ErrAlreadyExist пользователь уже существует.
	ErrAlreadyExist = errors.New("user already exist")
	// ErrNotFound пользователь не найден.
	ErrNotFound = errors.New("user not found")
	// ErrInvalidCredentials неверные данные.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrNoRowsUpdated не обновлено ни одной строки данных пользователя.
	ErrNoRowsUpdated = errors.New("none rows was updated")
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

// Register регистрация нового пользователя.
func (s *Service) Register(ctx context.Context, login, password, pepper string) (*User, error) {
	op := "domain.User.Register"

	var (
		user *User
		err  error
	)

	user, err = s.repo.GetByLogin(ctx, login)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("%s: failed to check existing of user %w", op, err)
	}
	if user != nil {
		return nil, ErrAlreadyExist
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

// Login авторизация пользователя.
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

// Update обновление данных о пользователе.
func (s *Service) Update(ctx context.Context, u *User) error {
	op := "domain.User.service.Update"

	var err error

	_, err = s.repo.GetByID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to check existing of user %w", op, err)
	}

	err = s.repo.Update(ctx, u)
	if err != nil {
		return fmt.Errorf("%s error updating user %w", op, err)
	}

	return nil
}
