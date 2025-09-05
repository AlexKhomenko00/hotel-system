-- name: InsertGuest :one
INSERT INTO booking.guests (id, first_name, last_name, email)
VALUES ($1, $2, $3, $4)
RETURNING id, first_name, last_name, email;
