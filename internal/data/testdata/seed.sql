
-- Seed data for users
INSERT INTO users (id, name, email, password_hash, activated, created_at) VALUES
( 1, 'Alice', 'alice@example.com', '$2a$12$QIlDOBLYkJ0QsAMOkv.taOzOT3FtgbR4ge.FpYpogLe4/Vv9N1s2e', true, '2024-01-01 10:00:00'),
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


-- Seed data for appointments
INSERT INTO appointments (id, creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence) VALUES
(1, 1, 2, 'Appointment 1', 'Description 1', '2023-06-01 12:00:00', '2023-06-01 13:00:00', 'Location 1', 'pending', '2023-06-01 10:00:00', '2023-06-01 10:00:00', 'UTC', 'public', 'daily'),
(2, 2, 1, 'Appointment 2', 'Description 2', '2023-06-02 14:00:00', '2023-06-02 15:00:00', 'Location 2', 'accepted', '2023-06-02 10:00:00', '2023-06-02 10:00:00', 'UTC', 'private', 'weekly');

-- Seed data for appointment_requests
INSERT INTO appointment_requests (request_id, requester_id, target_user_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone) VALUES
(1, 1, 2, 'Request 1', 'Description 1', '2023-06-01 12:00:00', '2023-06-01 13:00:00', 'Location 1', 'pending', '2023-06-01 10:00:00', '2023-06-01 10:00:00', 'UTC'),
(2, 1, 2, 'Request 2', 'Description 2', '2023-06-02 14:00:00', '2023-06-02 15:00:00', 'Location 2', 'accepted', '2023-06-02 10:00:00', '2023-06-02 10:00:00', 'UTC');

-- Seed data for appointment_events
INSERT INTO appointment_events (id, appointment_id, user_id, provider_name, provider_event_id) VALUES
(1, 1, 1, 'google', 'event_1'),
(2, 1, 2, 'outlook', 'event_2');

-- Adjust the sequences for all my tables
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users) + 1);
SELECT setval('groups_id_seq', (SELECT MAX(id) FROM groups) + 1);
SELECT setval('user_groups_id_seq', (SELECT MAX(id) FROM user_groups) + 1);
SELECT setval('appointments_id_seq', (SELECT MAX(id) FROM appointments) + 1);
SELECT setval('appointment_requests_request_id_seq', (SELECT MAX(request_id) FROM appointment_requests) + 1);
SELECT setval('appointment_events_id_seq', (SELECT MAX(id) FROM appointment_events) + 1);


