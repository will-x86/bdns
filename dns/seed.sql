-- test user
INSERT OR IGNORE INTO users (id, timezone) VALUES ('aabbccdd', 'Europe/London');
-- test profile
INSERT OR IGNORE INTO profiles (id, user_id, name) VALUES ('ppqqrrss', 'aabbccdd', 'test-laptop');


-- add dummy blocklist
INSERT OR IGNORE INTO blocklist_sources( id, name, url, category) VALUES ('ddccbbaa', 'test', 'https://example-list.com', 'test');
-- populate dummy blocklist
INSERT OR IGNORE INTO blocklist_entries( domain, source_id, category) VALUES ('example.com', 'ddccbbaa', 'test');
-- add fake blocks
INSERT OR IGNORE INTO user_category_blocks (profile_id, category) VALUES ('ppqqrrss', 'unified');
-- add fake blocks x2
INSERT OR IGNORE INTO user_category_blocks (profile_id, category) VALUES ('ppqqrrss', 'porn');

INSERT OR IGNORE INTO permanent_whitelists (profile_id, domain) VALUES ('ppqqrrss', 'ads.example.com');

INSERT OR IGNORE INTO temporary_whitelists (profile_id, domain, expires_at) VALUES ('ppqqrrss', 'tracker.example.com', unixepoch()+1000);

INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day) VALUES('ppqqrrss', 'test', 36, 48, 1);


-- Above should:
-- Create a profile with userid ppqqrrss
-- perm block porn + unified
-- perm whitelist ads.example.com
-- temp whitelist ads.example.com
