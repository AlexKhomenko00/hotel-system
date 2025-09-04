-- +goose Up
-- +goose StatementBegin
CREATE TABLE
	booking.reservations (
		id UUID PRIMARY KEY,
		hotel_id UUID NOT NULL,
		room_type_id UUID NOT NULL,
		start_date DATE NOT NULL,
		end_date DATE NOT NULL,
		status VARCHAR(50) NOT NULL,
		guest_id UUID NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (hotel_id) REFERENCES booking.hotels (id) ON DELETE CASCADE,
		FOREIGN KEY (room_type_id) REFERENCES booking.room_types (id) ON DELETE CASCADE,
		FOREIGN KEY (guest_id) REFERENCES booking.guests (id) ON DELETE CASCADE
	);

CREATE TABLE
	booking.room_type_inventory (
		hotel_id UUID NOT NULL,
		room_type_id UUID NOT NULL,
		date DATE NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		version BIGSERIAL NOT NULL,
		total_inventory INT NOT NULL,
		total_reserved INT NOT NULL DEFAULT 0,
		PRIMARY KEY (hotel_id, room_type_id, date),
		FOREIGN KEY (hotel_id) REFERENCES booking.hotels (id) ON DELETE CASCADE,
		FOREIGN KEY (room_type_id) REFERENCES booking.room_types (id) ON DELETE CASCADE
	);

-- +goose StatementEnd
-- 
-- +goose Down
-- +goose StatementBegin
DROP TABLE booking.reservations;

DROP TABLE booking.room_type_inventory;

-- +goose StatementEnd
