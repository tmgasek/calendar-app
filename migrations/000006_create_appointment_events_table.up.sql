CREATE TABLE appointment_events (
    id SERIAL PRIMARY KEY,
    appointment_id INT NOT NULL,
    CONSTRAINT fk_appointment_id FOREIGN KEY (appointment_id)
        REFERENCES appointments(id) ON DELETE CASCADE,
    user_id INT NOT NULL,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    provider_name VARCHAR(255) NOT NULL,
    provider_event_id VARCHAR(255) NOT NULL
);
