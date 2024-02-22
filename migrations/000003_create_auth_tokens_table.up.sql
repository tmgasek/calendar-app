CREATE TABLE auth_tokens (
    user_id INT,
    auth_provider TEXT,
    access_token TEXT,
    refresh_token TEXT,
    token_type TEXT,
    expiry TIMESTAMP,
    scope TEXT,
    PRIMARY KEY (user_id, auth_provider)
);
