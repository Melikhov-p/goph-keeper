package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/domain/secret"
)

type SecretRepository struct {
	db *sql.DB
}

func NewSecretRepository(db *sql.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (sr *SecretRepository) SaveSecret(ctx context.Context, s *secret.Secret) error {
	op := "repository.postgres.SaveSecret"

	var (
		tx    *sql.Tx
		query string
		err   error
	)

	tx, err = sr.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to start transaction for save secret %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	query = `
		INSERT INTO secrets (user_id, name, type, created_at, updated_at, version) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	row := tx.QueryRowContext(ctx, query, s.ID, s.Name, s.Type, s.CreatedAt, s.UpdatedAt, s.Version)
	if err = row.Scan(&s.ID); err != nil {
		return fmt.Errorf("%s: failed to query row for secret with error %w", op, err)
	}

	err = sr.saveSecretData(ctx, tx, s)
	if err != nil {
		return fmt.Errorf("%s: failed to save secret data with %w", op, err)
	}

	return nil
}

func (sr *SecretRepository) saveSecretData(ctx context.Context, tx *sql.Tx, s *secret.Secret) error {
	op := "repository.postgres.saveSecretData"

	var (
		query string
		err   error
	)

	switch s.Type {
	case secret.TypePassword:
		if data, ok := s.Data.(*secret.PasswordData); ok {
			query = `
					INSERT INTO password_data (secret_id, username, password_encrypted, url, notes_encrypted, metadata) 
					VALUES ($1, $2, $3, $4, $5, $6)
					`
			_, err = tx.ExecContext(ctx, query, s.ID, data.Username, data.Pass, data.URL, data.Notes, data.MetaData)
			break
		}
		return fmt.Errorf("%s: failed to assert secret data to PasswordData %w", op, err)
	case secret.TypeCard:
		if data, ok := s.Data.(*secret.CardData); ok {
			query = `
					INSERT INTO card_data (
										   secret_id, card_number_encrypted, card_holder_encrypted, 
										   expiry_date_encrypted, cvv_encrypted, notes_encrypted, metadata
										   )
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					`
			_, err = tx.ExecContext(
				ctx, query,
				s.ID, data.Number, data.Owner, data.ExpireDate, data.CVV, data.Notes, data.MetaData,
			)
			break
		}
		return fmt.Errorf("%s: failed to assert secret data to CardData %w", op, err)
	default:
		return fmt.Errorf("%s: invalid secret type %s", op, s.Type)
	}

	if err != nil {
		return fmt.Errorf(
			"%s: failed to exec context for secret data with type %s and error %w",
			op, s.Type, err,
		)
	}

	return nil
}
