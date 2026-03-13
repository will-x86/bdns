INSERT OR IGNORE INTO users (id, timezone) VALUES ('aabbccdd', 'Europe/London');
INSERT OR IGNORE INTO profiles (user_id, name) VALUES ('aabbccdd', 'test-laptop');

INSERT OR IGNORE INTO user_category_blocks (user_id, category) VALUES ('aabbccdd', 'unified');
INSERT OR IGNORE INTO user_category_blocks (user_id, category) VALUES ('aabbccdd', 'porn');

INSERT OR IGNORE INTO permanent_whitelists (user_id, domain) VALUES ('aabbccdd', 'ads.example.com');

INSERT OR IGNORE INTO temporary_whitelists (user_id, domain, expires_at) VALUES ('aabbccdd', 'tracker.example.com', 4070908800);
