package secret

import "context"

// Repository интерфейс для репозитория секретов.
type Repository interface {
	CreateSecret(ctx context.Context, secret *Secret)
	CreateSecretPassword(ctx context.Context, secret *Secret, data *PasswordData)
	CreateSecretCard(ctx context.Context, secret *Secret, data *CardData)
	CreateSecretFile(ctx context.Context, secret *Secret, data *FileData)
}
