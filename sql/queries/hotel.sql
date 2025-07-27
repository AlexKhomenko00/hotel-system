-- name: CreateHotel :one
INSERT INTO
	booking.hotels (id, name, location)
VALUES
	($1, $2, $3)
RETURNING
	*;

-- name: GetHotelById :one
SELECT
	*
FROM
	booking.hotels
WHERE
	id = $1;

-- name: GetHotelByName :one
SELECT
	*
FROM
	booking.hotels
WHERE
	name = $1;

-- name: UpdateHotel :one
UPDATE booking.hotels
SET
	name = $2,
	location = $3,
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = $1
RETURNING *;

-- name: DeleteHotel :exec
UPDATE booking.hotels
SET
	deleted_at = CURRENT_TIMESTAMP
WHERE
	id = $1;
