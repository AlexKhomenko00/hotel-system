-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA booking;

CREATE TABLE
	booking.hotels (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
		name VARCHAR(255) NOT NULL UNIQUE,
		location VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
	);

CREATE TABLE
	booking.room_types (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
		hotel_id UUID NOT NULL REFERENCES booking.hotels (id),
		name VARCHAR(255) NOT NULL,
		description TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (hotel_id, name)
	);

CREATE TYPE booking.room_status AS ENUM('available', 'unavailable', 'maintenance');

CREATE TABLE
	booking.rooms (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
		hotel_id UUID NOT NULL REFERENCES booking.hotels (id),
		room_type_id UUID NOT NULL REFERENCES booking.room_types (id),
		status booking.room_status NOT NULL DEFAULT 'available',
		name VARCHAR(255) NOT NULL,
		description TEXT,
		number INT NOT NULL,
		floor INT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (hotel_id, number)
	);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE booking.rooms;

DROP TABLE booking.room_types;

DROP TABLE booking.hotels;

DROP SCHEMA booking;

-- +goose StatementEnd
