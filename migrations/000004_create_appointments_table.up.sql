CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
        CONSTRAINT fk_user_id FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    google_event_id VARCHAR(255),
    microsoft_event_id VARCHAR(255),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location VARCHAR(255),
    status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    time_zone VARCHAR(100),
    visibility VARCHAR(50),
    recurrence TEXT
);
