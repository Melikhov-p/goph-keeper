package secret

import "context"

// Repository интерфейс для репозитория секретов.
type Repository interface {
	Create(ctx context.Context, secret *Secret)
	CreatePassword(ctx context.Context, secret *Secret)
}
