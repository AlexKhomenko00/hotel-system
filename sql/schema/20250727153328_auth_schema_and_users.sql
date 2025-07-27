-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA auth;

CREATE TABLE
	auth.users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
	);

CREATE UNIQUE INDEX idx_users_email ON auth.users (TRIM(LOWER(email)));

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE auth.users;

DROP SCHEMA auth;

-- +goose StatementEnd
