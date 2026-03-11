INSERT INTO users (id, timezone) VALUES ('aabbccdd', 'Europe/London');
INSERT INTO profiles (user_id, name) VALUES ('aabbccdd', 'test-laptop');

INSERT INTO user_category_blocks (user_id, category) VALUES ('aabbccdd', 'unified');
INSERT INTO user_category_blocks (user_id, category) VALUES ('aabbccdd', 'porn');

INSERT INTO permanent_whitelists (user_id, domain) VALUES ('aabbccdd', 'ads.example.com');

INSERT INTO temporary_whitelists (user_id, domain, expires_at) VALUES ('aabbccdd', 'tracker.example.com', 4070908800);
