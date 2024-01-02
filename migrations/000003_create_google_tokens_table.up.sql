CREATE TABLE google_tokens (
    user_id INT PRIMARY KEY,
    access_token TEXT,
    refresh_token TEXT,
    token_type TEXT,
    expiry TIMESTAMP,
    scope TEXT
);
