-- +goose Up
-- +goose StatementBegin
CREATE TABLE blocklist_sources (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    name       TEXT NOT NULL,           -- "IOSD, etc"
    url        TEXT,                    -- null if local/manual
    category   TEXT NOT NULL,           -- "ads", "adult", "social", etc.
    enabled    INTEGER NOT NULL DEFAULT 1,
    last_synced_at INTEGER,
    entry_count INTEGER DEFAULT 0
);
CREATE TABLE blocklist_entries (
    domain     TEXT NOT NULL,
    source_id  TEXT NOT NULL REFERENCES blocklist_sources(id) ON DELETE CASCADE,
    category   TEXT NOT NULL,
    PRIMARY KEY (domain, source_id)
);
CREATE INDEX idx_blocklist_domain ON blocklist_entries(domain);

-- Per profile block rules
CREATE TABLE block_rules (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    profile_id TEXT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    domain     TEXT,
    category   TEXT,
    action     TEXT NOT NULL CHECK(action IN ('block','allow')),
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- A rule with NO entries in this table = always active
-- A rule WITH entries = only active during those windows - active = block at those times
CREATE TABLE block_rule_schedules (
    id       TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    rule_id  TEXT NOT NULL REFERENCES block_rules(id) ON DELETE CASCADE,
    days     TEXT NOT NULL,   -- "mon,wed,fri" or "everyday" or "weekdays"
    time_start TEXT NOT NULL, -- "09:00"
    time_end   TEXT NOT NULL  -- "17:00"
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS block_rule_schedules;
DROP TABLE IF EXISTS block_rules;
DROP TABLE IF EXISTS blocklist_entries;
DROP TABLE IF EXISTS blocklist_sources;
-- +goose StatementEnd
