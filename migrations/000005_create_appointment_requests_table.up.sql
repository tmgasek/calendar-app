CREATE TABLE appointment_requests (
    request_id SERIAL PRIMARY KEY,
    requester_id INT NOT NULL,
        CONSTRAINT fk_requester_id FOREIGN KEY (requester_id)
        REFERENCES users(id) ON DELETE CASCADE,
    target_user_id INT NOT NULL,
        CONSTRAINT fk_target_user_id FOREIGN KEY (target_user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',  -- Could be 'pending', 'accepted', or 'declined'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    time_zone VARCHAR(100)
);
