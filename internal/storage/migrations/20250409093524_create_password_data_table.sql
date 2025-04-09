-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS password_data (
                               secret_id INTEGER PRIMARY KEY REFERENCES secrets(id) ON DELETE CASCADE,
                               username TEXT,
                               password_encrypted TEXT NOT NULL,
                               url TEXT,
                               notes_encrypted TEXT,
                               metadata JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS password_data;
-- +goose StatementEnd
