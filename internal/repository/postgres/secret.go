package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/domain/secret"
	"github.com/Melikhov-p/goph-keeper/internal/repository/external_storage"
	"go.uber.org/zap"
)

var errInvalidJSON = errors.New("invalid json meta data")

// SecretRepository репозиторий секретов.
type SecretRepository struct {
	db  *sql.DB
	log *zap.Logger
}

// NewSecretRepository новый репозиторий для секретов.
func NewSecretRepository(db *sql.DB, l *zap.Logger) *SecretRepository {
	return &SecretRepository{db: db, log: l}
}

// SaveSecret сохранить секрет в БД.
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

	row := tx.QueryRowContext(ctx, query, s.UserID, s.Name, s.Type, s.CreatedAt, s.UpdatedAt, s.Version)
	if err = row.Scan(&s.ID); err != nil {
		return fmt.Errorf("%s: failed to query row for secret with error %w", op, err)
	}

	err = sr.saveSecretData(ctx, tx, s)
	if err != nil {
		return fmt.Errorf("%s: failed to save secret data with %w", op, err)
	}

	return nil
}

// saveSecretData сохранить секретные данные в БД.
func (sr *SecretRepository) saveSecretData(ctx context.Context, tx *sql.Tx, s *secret.Secret) error {
	op := "repository.postgres.saveSecretData"

	var (
		query string
		err   error
	)

	switch s.Type {
	case secret.TypePassword:
		sr.log.Debug("new password secret", zap.Any("SECRET", s))
		if data, ok := s.Data.(*secret.PasswordData); ok {
			if data.MetaData == nil {
				data.MetaData = []byte("{}")
			}
			if data.MetaData != nil && !json.Valid(data.MetaData) {
				return errInvalidJSON
			}
			query = `
					INSERT INTO password_data (secret_id, username, password_encrypted, url, notes_encrypted, metadata) 
					VALUES ($1, $2, $3, $4, $5, $6)
					`
			_, err = tx.ExecContext(
				ctx, query, s.ID, data.Username, data.Pass, data.URL, data.Notes,
				data.MetaData)
			break
		}
		return fmt.Errorf("%s: failed to assert secret data to PasswordData %w", op, err)
	case secret.TypeCard:
		if data, ok := s.Data.(*secret.CardData); ok {
			if data.MetaData == nil {
				data.MetaData = []byte("{}")
			}
			if data.MetaData != nil && !json.Valid(data.MetaData) {
				return errInvalidJSON
			}
			query = `
					INSERT INTO card_data (
										   secret_id, card_number_encrypted, card_holder_encrypted, 
										   expiry_date_encrypted, cvv_encrypted, notes_encrypted, metadata
										   )
					VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
					`
			_, err = tx.ExecContext(
				ctx, query,
				s.ID, data.Number, data.Owner, data.ExpireDate, data.CVV,
				data.Notes, data.MetaData,
			)
			break
		}
		return fmt.Errorf("%s: failed to assert secret data to CardData %w", op, err)
	case secret.TypeBinary:
		if data, ok := s.Data.(*secret.FileData); ok {
			if data.MetaData == nil {
				data.MetaData = []byte("{}")
			}
			if data.MetaData != nil && !json.Valid(data.MetaData) {
				return errInvalidJSON
			}

			var checksum string
			checksum, data.Path, err = external_storage.SaveFileData(ctx, s.UserID, data.Path, data.Content)
			if err != nil {
				return fmt.Errorf("%s: failed to save binary content to file with error %w", op, err)
			}

			query = `
					INSERT INTO external_storage (secret_id, storage_path, storage_type, filename, checksum, created_at) 
					VALUES ($1, $2, $3, $4, $5, $6)
					`
			_, err = tx.ExecContext(ctx, query, s.ID, data.Path, "note", data.Name, checksum, s.CreatedAt)
			break
		}
		return fmt.Errorf("%s: failed to assert secret data to FileData %w", op, err)
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

// GetSecretsByName получить секреты с заданным названием.
func (sr *SecretRepository) GetSecretsByName(
	ctx context.Context,
	secretName string,
	userID int,
) ([]*secret.Secret, error) {
	op := "repository.postgres.GetSecretsByName"

	var (
		rows *sql.Rows
		err  error
	)
	secrets := make([]*secret.Secret, 0)

	query := `
		SELECT id, user_id, name, type, created_at, updated_at, version 
		FROM secrets WHERE name = $1 AND user_id = $2
	`

	rows, err = sr.db.QueryContext(ctx, query, secretName, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, secret.ErrSecretNotFound
		}
		return nil, fmt.Errorf("%s: failed to query context with error %w", op, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var s secret.Secret
		if err = rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Type, &s.CreatedAt, &s.UpdatedAt, &s.Version); err != nil {
			return nil, fmt.Errorf("%s: failed to scan row for secret with error %w", op, err)
		}

		secrets = append(secrets, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: got rows.Err: %w", op, err)
	}

	for _, s := range secrets {
		switch s.Type {
		case secret.TypePassword:
			query = `SELECT username, password_encrypted, url, notes_encrypted, metadata 
					FROM password_data WHERE secret_id = $1`
		case secret.TypeCard:
			query = `SELECT card_number_encrypted, card_holder_encrypted, 
							expiry_date_encrypted, cvv_encrypted, notes_encrypted, metadata
					FROM card_data WHERE secret_id = $1`
		case secret.TypeBinary:
			query = `
					SELECT storage_path, filename
					FROM external_storage WHERE secret_id = $1
					`
		default:
			return nil, fmt.Errorf("%s: invalid secret type %s", op, s.Type)
		}

		row := sr.db.QueryRowContext(ctx, query, s.ID)
		if err = s.SetDataFromRow(row); err != nil {
			return nil, fmt.Errorf("%s: failed to set data from row with error %w", op, err)
		}

		if err = row.Err(); err != nil {
			return nil, fmt.Errorf("%s: got row.Err: %w", op, err)
		}
	}

	return secrets, nil
}
