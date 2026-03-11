-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    timezone TEXT NOT NULL,        -- "Europe/London"
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE TABLE profiles (
    id   TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,         -- "laptop", "work phone"
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
