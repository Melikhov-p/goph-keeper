package user

import "context"

// Repository интерфейс репозитория для пользователя.
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}
