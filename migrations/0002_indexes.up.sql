CREATE INDEX slots_room_date_idx ON slots (room_id, start_at);

CREATE UNIQUE INDEX bookings_one_active_per_slot_idx
ON bookings (slot_id)
WHERE status_id  = 1;