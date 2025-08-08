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

-- name: BatchUpdateRoomTypeInventory :execrows
INSERT INTO
	booking.room_type_inventory (
		hotel_id,
		room_type_id,
		date,
		total_inventory,
		total_reserved,
		updated_at,
		created_at
	)
SELECT
	@hotel_id,
	@room_type_id,
	unnest(@dates),
	@total_inventory,
	0,
	CURRENT_TIMESTAMP,
	CURRENT_TIMESTAMP
	--  simplification of a business rule to not handle constantly changing overcapacity issue-> e.g. some rolling update in the future is possible
	ON CONFLICT (hotel_id, room_type_id, date)
DO NOTHING;
