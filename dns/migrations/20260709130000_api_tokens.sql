-- +goose Up
-- +goose StatementBegin
-- Per-user API token. SQLite can't default ADD COLUMN to randomblob, so
-- backfill here; the app sets a token on new inserts.
ALTER TABLE users ADD COLUMN api_token TEXT;
UPDATE users SET api_token = lower(hex(randomblob(16))) WHERE api_token IS NULL;
CREATE UNIQUE INDEX idx_users_api_token ON users(api_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_api_token;
ALTER TABLE users DROP COLUMN api_token;
-- +goose StatementEnd
