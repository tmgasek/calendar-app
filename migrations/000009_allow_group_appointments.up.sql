-- Add group_id to appointment_requests
ALTER TABLE appointment_requests
ADD COLUMN group_id INT,
ADD CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

-- Add group_id to appointments
ALTER TABLE appointments
ADD COLUMN group_id INT,
ADD CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

-- Add appointment_type to appointment_requests. Can be 'individual' or 'group'
ALTER TABLE appointment_requests
ADD COLUMN appointment_type VARCHAR(20) DEFAULT 'individual';

-- Add appointment_type to appointments. Can be 'individual' or 'group'
ALTER TABLE appointments
ADD COLUMN appointment_type VARCHAR(20) DEFAULT 'individual';
