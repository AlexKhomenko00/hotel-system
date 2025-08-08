-- +goose Up
-- +goose StatementBegin
ALTER TABLE booking.hotels
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE booking.hotels
DROP COLUMN is_active;

-- +goose StatementEnd
