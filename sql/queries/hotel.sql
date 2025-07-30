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
RETURNING
	*;

-- name: DeleteHotel :exec
UPDATE booking.hotels
SET
	deleted_at = CURRENT_TIMESTAMP
WHERE
	id = $1;

-- name: FindRoomTypesByHotelIdAndName :one
SELECT
	*
FROM
	booking.room_types
WHERE
	hotel_id = $1
	AND name = $2;

-- name: CreateRoomType :one
INSERT INTO
	booking.room_types (id, hotel_id, name, description)
VALUES
	($1, $2, $3, $4)
RETURNING
	*;

-- name: GetRoomTypeByIdAndHotelId :one
SELECT
	*
FROM
	booking.room_types
WHERE
	id = $1
	AND hotel_id = $2;

-- name: UpdateRoomType :one
UPDATE booking.room_types
SET
	name = $3,
	description = $4,
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = $1
	AND hotel_id = $2
RETURNING
	*;

-- name: DeleteRoomType :exec
UPDATE booking.room_types
SET
	deleted_at = CURRENT_TIMESTAMP
WHERE
	id = $1
	AND hotel_id = $2;
