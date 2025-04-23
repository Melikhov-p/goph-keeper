package secret

import "context"

// Repository интерфейс для репозитория секретов.
type Repository interface {
	// SaveSecret сохраняет новый секрет в базу данных возвращая ошибку или её отсутствие.
	SaveSecret(ctx context.Context, secret *Secret) error
	// GetSecretsByName поиск среди всех секретов по названию секрета.
	GetSecretsByName(ctx context.Context, secretName string, userID int) ([]*Secret, error)
}
