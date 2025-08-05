-- name: GetReservationById :one
SELECT
	*
FROM
	booking.reservations
WHERE
	id = $1;

-- name: InsertReservation :one
INSERT INTO
	booking.reservations (
		id,
		hotel_id,
		room_type_id,
		start_date,
		end_date,
		status,
		guest_id,
		updated_at,
		created_at
	)
VALUES
	(
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		CURRENT_TIMESTAMP,
		CURRENT_TIMESTAMP
	)
RETURNING
	id;
