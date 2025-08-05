-- +goose Up
-- +goose StatementBegin
CREATE TABLE
	booking.guests (
		id UUID PRIMARY KEY,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE
	);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE booking.guests;

-- +goose StatementEnd
