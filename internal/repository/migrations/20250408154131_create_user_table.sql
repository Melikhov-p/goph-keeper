-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       login TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_login ON users(login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP INDEX IF EXISTS idx_users_login;
-- +goose StatementEnd
