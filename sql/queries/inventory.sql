-- name: GetHotelInventoryForRange :many
SELECT
	*
FROM
	booking.room_type_inventory
WHERE
	room_type_id = $1
	AND hotel_id = $2
	AND date BETWEEN $3 AND $4;

-- name: UpdateRoomTypeInventoryForDate :execrows
UPDATE booking.room_type_inventory
SET
	total_reserved = total_reserved + $1,
	version = version + 1,
	updated_at = CURRENT_TIMESTAMP
WHERE
	room_type_id = $2
	AND hotel_id = $3
	AND date = $4
	AND version = $5;
