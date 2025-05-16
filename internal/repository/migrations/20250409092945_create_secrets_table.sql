-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS secrets (
                         id SERIAL PRIMARY KEY,
                         user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         name TEXT NOT NULL,
                         type TEXT NOT NULL CHECK ( type IN ('password', 'card', 'note', 'binary') ),
                         created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                         deleted_at TIMESTAMP WITH TIME ZONE,
                         version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_secrets_user_id ON secrets(user_id);
CREATE INDEX idx_secrets_type ON secrets(type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS secret_type;
DROP TABLE IF EXISTS secrets;
DROP INDEX IF EXISTS idx_secrets_user_id;
DROP INDEX IF EXISTS idx_secrets_type;
-- +goose StatementEnd
