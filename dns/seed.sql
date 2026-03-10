INSERT INTO users (id) VALUES ('aabbccdd');
INSERT INTO profiles (user_id, name) VALUES ('aabbccdd', 'my profile');


INSERT INTO blocklist_sources (
    id, name, url, category, enable, last_synced_at
    ) VALUES (
    'abcdabcd', 'Stevenblack', 'https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/social-only/hosts', 
    'social', true, 0);
