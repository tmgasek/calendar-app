
-- Seed data for users
INSERT INTO users (id, name, email, password_hash, activated, created_at) VALUES
( 1, 'Alice Jones', 'alice@example.com', '$2a$12$QIlDOBLYkJ0QsAMOkv.taOzOT3FtgbR4ge.FpYpogLe4/Vv9N1s2e', true, '2024-01-01 10:00:00'),
(2, 'Bob', 'bob@example.com', 'password_hash_2', true, '2023-06-01 11:00:00'),
(3, 'Charlie', 'charlie@example.com', 'password_hash_3', true, '2023-06-01 12:00:00');

-- Seed data for groups
INSERT INTO groups (id, name, description, created_at, updated_at) VALUES
(1, 'Group 1', 'Description for Group 1', '2023-06-01 13:00:00', '2023-06-01 13:00:00'),
(2, 'Group 2', 'Description for Group 2', '2023-06-01 14:00:00', '2023-06-01 14:00:00');

-- Seed data for user_groups
INSERT INTO user_groups (user_id, group_id) VALUES
(1, 1),
(2, 1),
(1, 2);

INSERT INTO auth_tokens (user_id, auth_provider, access_token, refresh_token, token_type, expiry, scope) VALUES
(1, 'google', 'access-token-1', 'refresh-token-1', 'Bearer', NOW() + INTERVAL '1 hour', 'scope-1');

-- Adjust the sequences for all my tables
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users) + 1);
SELECT setval('groups_id_seq', (SELECT MAX(id) FROM groups) + 1);
SELECT setval('user_groups_id_seq', (SELECT MAX(id) FROM user_groups) + 1);
