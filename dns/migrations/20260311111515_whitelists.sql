-- +goose Up
-- +goose StatementBegin
CREATE TABLE permanent_whitelists (
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    domain     TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (profile_id, domain)
);

CREATE TABLE temporary_whitelists (
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    domain     TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (profile_id, domain)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS temporary_whitelists;
DROP TABLE IF EXISTS permanent_whitelists;
-- +goose StatementEnd
