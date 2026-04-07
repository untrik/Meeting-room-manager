CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE roles (
    id SMALLINT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);

CREATE TABLE booking_statuses (
    id SMALLINT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    role_id SMALLINT NOT NULL REFERENCES roles(id),
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    capacity INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week SMALLINT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    CONSTRAINT schedules_room_unique UNIQUE (room_id),
    CONSTRAINT schedules_time_valid CHECK (start_time < end_time)
);

CREATE TABLE slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT slots_time_valid CHECK (start_at < end_at),
    CONSTRAINT slots_unique UNIQUE (room_id, start_at, end_at)
);
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE RESTRICT,
    status_id SMALLINT NOT NULL REFERENCES booking_statuses(id),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);