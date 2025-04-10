// Package postgres пакет с имплементацией интерфейсов репозиториев через БД Postgres.
package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/goph-keeper/internal/config"
	_ "github.com/jackc/pgx" // Импорт драйвера для postgres.
	"github.com/pressly/goose"
)

// NewConnection установка соединения с базой данных.
func NewConnection(cfg *config.Config) (*sql.DB, error) {
	op := "repository.Postgres.ConnectDB"

	db, err := sql.Open("pgx", cfg.Database.URI)
	if err != nil {
		return nil, fmt.Errorf("%s: error open sql connection to database: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s error ping database %w", op, err)
	}

	if err = makeMigrations(cfg, db); err != nil {
		return nil, fmt.Errorf("%s: error making migrations %w", op, err)
	}

	return db, nil
}

// makeMigrations выполнение миграций через goose.
func makeMigrations(cfg *config.Config, db *sql.DB) error {
	op := "repository.Potsgres.makeMigrations"

	var err error

	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: error goose set dialect %w", op, err)
	}

	if err = goose.Up(db, cfg.Database.MigrationsPath); err != nil {
		return fmt.Errorf("%s: error up migrations %w", op, err)
	}

	return nil
}
