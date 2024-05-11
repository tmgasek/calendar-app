-- Remove appointment_type from appointments
ALTER TABLE appointments
DROP COLUMN appointment_type;

-- Remove appointment_type from appointment_requests
ALTER TABLE appointment_requests
DROP COLUMN appointment_type;

-- Remove group_id from appointments
ALTER TABLE appointments
DROP CONSTRAINT fk_group_id,
DROP COLUMN group_id;

-- Remove group_id from appointment_requests
ALTER TABLE appointment_requests
DROP CONSTRAINT fk_group_id,
DROP COLUMN group_id;
