DROP INDEX IF EXISTS bookings_user_status_idx;
DROP INDEX IF EXISTS bookings_one_active_per_slot;
DROP INDEX IF EXISTS slots_room_start_idx;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;