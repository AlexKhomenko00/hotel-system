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
	unnest(@dates::timestamp[]),
	@total_inventory,
	0,
	CURRENT_TIMESTAMP,
	CURRENT_TIMESTAMP
	--  simplification of a business rule to not handle constantly changing overcapacity issue-> e.g. some rolling update in the future is possible
	ON CONFLICT (hotel_id, room_type_id, date)
DO NOTHING;

-- name: GetRoomAvailabilityByDates :many
SELECT 
	rti.room_type_id,
	rt.name as room_type_name,
	rt.description,
	rti.date,
	(rti.total_inventory - rti.total_reserved) as available_capacity
FROM booking.room_type_inventory rti
INNER JOIN booking.room_types rt ON rti.room_type_id = rt.id
WHERE rti.hotel_id = @hotel_id
	AND rti.date BETWEEN @check_in AND @check_out
	AND rti.total_reserved < ((@overbooking)::float * rti.total_inventory)
ORDER BY rt.name, rti.date;
