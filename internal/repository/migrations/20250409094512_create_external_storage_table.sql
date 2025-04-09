-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS external_storage (
                                  id SERIAL PRIMARY KEY,
                                  secret_id INTEGER NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
                                  storage_path TEXT NOT NULL,  -- Путь к файлу (относительный или полный)
                                  storage_type TEXT NOT NULL CHECK (storage_type IN ('note', 'binary')),
                                  checksum TEXT,  -- Контрольная сумма для проверки целостности
                                  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_external_storage_secret ON external_storage(secret_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS external_storage;
DROP INDEX IF EXISTS idx_external_storage_secret;
-- +goose StatementEnd
