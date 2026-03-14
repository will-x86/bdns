-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_category_blocks (
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    category   TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (profile_id, category)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_category_blocks;
-- +goose StatementEnd
