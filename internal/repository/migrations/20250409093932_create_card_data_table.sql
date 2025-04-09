-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS card_data (
                           secret_id INTEGER PRIMARY KEY REFERENCES secrets(id) ON DELETE CASCADE,
                           card_number_encrypted TEXT NOT NULL,
                           card_holder_encrypted TEXT,
                           expiry_date_encrypted TEXT,
                           cvv_encrypted TEXT,
                           notes_encrypted TEXT,
                           metadata JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS card_data;
-- +goose StatementEnd
