-- name: GetUserByEmail :one
SELECT
	*
FROM
	auth.users
WHERE
	email = $1;

-- name: GetUserById :one
SELECT
	*
FROM
	auth.users
WHERE
	id = $1
LIMIT
	1;

-- name: InsertUser :one
INSERT INTO
	auth.users (id, email, password_hash, guest_id)
VALUES
	($1, $2, $3, $4)
RETURNING
	*;
