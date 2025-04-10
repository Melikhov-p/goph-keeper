-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_keys (
                           user_id INT PRIMARY KEY REFERENCES users(id),
                           encrypted_key BYTEA NOT NULL,
                           created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_keys;
-- +goose StatementEnd
