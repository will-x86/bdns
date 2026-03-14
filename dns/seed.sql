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



-- test timeblocks
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 1, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 2, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 3, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 4, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 5, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 6, 0);
INSERT OR IGNORE INTO user_time_blocks (profile_id, category, start_time, end_time, day, created_at) VALUES('ppqqrrss', 'test', 0, 95, 7, 0);


-- Above should:
-- Create a profile with userid ppqqrrss
-- perm block porn + unified
-- perm whitelist ads.example.com
-- temp whitelist tracker.example.com
-- add category block from 
-- block category "test" from mid-day till 12:00 ( 15 min blocks) every day
