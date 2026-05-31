-- name: CreateUser :one
INSERT INTO users (username, full_name, email, hashed_password) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUser :one
SELECT * FROM users  WHERE (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')::uuid)
                       AND (sqlc.narg('username')::varchar IS NULL OR username=sqlc.narg('username')::varchar)
                       AND (sqlc.narg('email')::varchar IS NULL OR email=sqlc.narg('email')::varchar);

-- name: UserExists :one
SELECT EXISTS(SELECT 1
              FROM users
              WHERE (sqlc.narg('id')::uuid IS NULL OR id = sqlc.narg('id')::uuid)
              AND (sqlc.narg('username')::varchar IS NULL OR username=sqlc.narg('username')::varchar)
              AND (sqlc.narg('email')::varchar IS NULL OR email=sqlc.narg('email')::varchar)
              AND (hashed_password=sqlc.arg('password')::varchar));

-- name: UpdatePassword :one
UPDATE users
SET hashed_password = sqlc.arg('new_password'),
    updated_at = now()
WHERE (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
AND (sqlc.narg('username')::varchar IS NULL OR username=sqlc.narg('username')::varchar)
RETURNING *;


