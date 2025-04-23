package secret

import "context"

// Repository интерфейс для репозитория секретов.
type Repository interface {
	// SaveSecret сохраняет новый секрет в базу данных возвращая его ID или ошибку.
	SaveSecret(ctx context.Context, secret *Secret) (int, error)
	GetSecretsByName(ctx context.Context, secretName string, userID int) ([]*Secret, error)
}
