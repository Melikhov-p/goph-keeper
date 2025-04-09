package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

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
