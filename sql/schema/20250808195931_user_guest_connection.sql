-- +goose Up
-- +goose StatementBegin
-- Breaking change normally not allowed on production :)
ALTER TABLE auth.users
ADD COLUMN guest_id UUID NOT NULL REFERENCES booking.guests (id) UNIQUE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth.users
DROP COLUMN guest_id;

-- +goose StatementEnd
