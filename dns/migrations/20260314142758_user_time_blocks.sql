-- +goose Up
-- +goose StatementBegin
CREATE table user_time_blocks(
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    start_time INTEGER NOT NULL CHECK (start_time >= 0 AND start_time<= 95), -- 15 min blocks
    end_time INTEGER NOT NULL CHECK (end_time>= 0 AND end_time<= 95), -- inclusive
    -- user with start_time: 0 & end_time 3 cannot use during 00:00 -> 1:01
    day INTEGER NOT NULL CHECK (day >=0 AND day <=7),
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_time_blocks;
-- +goose StatementEnd
