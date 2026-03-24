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


-- add user with no blocks

INSERT OR IGNORE INTO users (id, timezone) VALUES ('noblock', 'Europe/London');
INSERT OR IGNORE INTO profiles (id, user_id, name) VALUES ('noblock', 'noblock', 'test-no-block');


-- users for 100 on social medai pool
INSERT OR IGNORE INTO users(id, timezone) VALUES ('user1', 'Europe/London');
INSERT OR IGNORE INTO users(id, timezone) VALUES ('user2', 'Europe/London');
INSERT OR IGNORE INTO users(id, timezone) VALUES ('user3', 'Europe/London');

-- Friendships (bidirectional)
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user1', 'user2');
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user2', 'user1');
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user1', 'user3');
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user3', 'user1');
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user2', 'user3');
INSERT OR IGNORE INTO user_friends(user_id, friend_id) VALUES ('user3', 'user2');

-- One profile per user
INSERT OR IGNORE INTO profiles(id, user_id, name) VALUES ('profile1', 'user1', 'User 1 Personal');
INSERT OR IGNORE INTO profiles(id, user_id, name) VALUES ('profile2', 'user2', 'User 2 Personal');
INSERT OR IGNORE INTO profiles(id, user_id, name) VALUES ('profile3', 'user3', 'User 3 Personal');

-- da poiol
INSERT OR IGNORE INTO friend_pools(id, created_by, name, pool_mode, total_limit)
VALUES ('pool1', 'user1', 'Social Media Shared Pot', 'shared', 100);

-- Add profiles in the pool
INSERT OR IGNORE INTO friend_pool_members(pool_id, profile_id) VALUES ('pool1', 'profile1');
INSERT OR IGNORE INTO friend_pool_members(pool_id, profile_id) VALUES ('pool1', 'profile2');
INSERT OR IGNORE INTO friend_pool_members(pool_id, profile_id) VALUES ('pool1', 'profile3');

INSERT OR IGNORE INTO friend_pool_category_blocks(pool_id, category) VALUES ('pool1', 'social');
