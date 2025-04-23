package secret

import (
	"context"
)

// Repository интерфейс для репозитория секретов.
type Repository interface {
	CreateSecretPassword(ctx context.Context) error
	CreateSecretCard(ctx context.Context) error
	CreateSecretFile(ctx context.Context) error
}
