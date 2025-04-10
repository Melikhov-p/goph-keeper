package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
)

// UserRepository репозиторий пользователя.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository получение репозитория пользователя.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создание нового пользователя.
func (ur *UserRepository) Create(ctx context.Context, u *user.User) error {
	op := "repository.Postgres.User.Create"

	query := `
		INSERT INTO users (login, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id
        `

	row := ur.db.QueryRowContext(ctx, query, u.Login, u.PassHash, u.CreatedAt, u.UpdatedAt)
	if err := row.Scan(&u.ID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetByID получение пользователя по ID.
func (ur *UserRepository) GetByID(ctx context.Context, id int) (*user.User, error) {
	op := "repository.Postgres.User.GetByID"

	var u user.User

	query := `SELECT (login, password_hash, created_at, updated_at) from users WHERE id = $1`

	row := ur.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(u.Login, u.PassHash, u.CreatedAt, u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("%s: error scanning row for user (%d) %w", op, id, err)
	}
	u.ID = id

	return &u, nil
}

// GetByLogin получение пользователя по логину.
func (ur *UserRepository) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	op := "repository.Postgres.User.GetByLogin"

	var u user.User

	query := `SELECT (id, password_hash, created_at, updated_at) from users WHERE login = $1`

	row := ur.db.QueryRowContext(ctx, query, login)
	if err := row.Scan(u.ID, u.PassHash, u.CreatedAt, u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("%s: error scanning row for user (%d) %w", op, login, err)
	}
	u.Login = login

	return &u, nil
}

// Update обновление информации пользователя.
func (ur *UserRepository) Update(ctx context.Context, u *user.User) error {
	op := "repository.Postgres.User.Update"

	query := `
	UPDATE users SET
	                 login = $1,
	                 updated_at = $2,
	                 password_hash = $3
	WHERE id = $4
	`

	res, err := ur.db.ExecContext(ctx, query, u.Login, u.UpdatedAt, u.PassHash, u.ID)
	if err != nil {
		return fmt.Errorf("%s: error executing context for update user %w", op, err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return user.ErrNoRowsUpdated
	}

	return nil
}

// Delete удаление пользователя.
func (ur *UserRepository) Delete(ctx context.Context, id int) error {
	op := "repository.Postgres.User.Delete"

	query := `DELETE FROM users WHERE id = $1`

	_, err := ur.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: error executing context for delete %w", op, err)
	}

	return nil
}
