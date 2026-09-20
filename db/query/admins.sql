-- name: IsAdmin :one
SELECT EXISTS(SELECT 1 FROM admins WHERE user_id = $1::uuid) AS is_admin;

-- name: AddAdmin :exec
INSERT INTO admins (user_id) VALUES ($1::uuid);

-- name: RemoveAdmin :exec
DELETE FROM admins WHERE user_id = $1::uuid;
