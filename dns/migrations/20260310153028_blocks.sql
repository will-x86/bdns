-- +goose Up
-- +goose StatementBegin
CREATE TABLE blocklist_sources (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4)))),
    name       TEXT NOT NULL,           -- "IOSD, etc"
    url        TEXT,                    -- null if local/manual
    category   TEXT NOT NULL,           -- "ads", "adult", "social", etc.
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS block_rule_schedules;
DROP TABLE IF EXISTS block_rules;
DROP TABLE IF EXISTS blocklist_entries;
DROP TABLE IF EXISTS blocklist_sources;
-- +goose StatementEnd
