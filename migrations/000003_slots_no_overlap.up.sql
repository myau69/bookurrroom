CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE slots
    ADD CONSTRAINT slots_room_time_no_overlap
    EXCLUDE USING gist (
    room_id WITH =,
    tstzrange(start_utc, end_utc, '[)') WITH &&
);