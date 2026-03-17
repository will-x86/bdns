-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_friends (
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id != friend_id)
);
CREATE TABLE friend_pools (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    created_by  TEXT REFERENCES users(id) ON DELETE SET NULL,
    name        TEXT,
    pool_mode   TEXT NOT NULL CHECK (pool_mode IN ('shared', 'borrow')),
    total_limit INTEGER NOT NULL DEFAULT 6000, -- shared = pot, borrow = per-member limit
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE TABLE friend_pool_members (
    pool_id    TEXT NOT NULL REFERENCES friend_pools(id) ON DELETE CASCADE,
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    joined_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (pool_id, profile_id)
);
CREATE TABLE friend_pool_category_blocks (
    pool_id    TEXT NOT NULL REFERENCES friend_pools(id) ON DELETE CASCADE,
    category   TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (pool_id, category)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS friend_pool_invites;
DROP TABLE IF EXISTS friend_pool_category_blocks;
DROP TABLE IF EXISTS friend_pool_members;
DROP TABLE IF EXISTS friend_pools;
DROP TABLE IF EXISTS user_friends;
-- +goose StatementEnd

