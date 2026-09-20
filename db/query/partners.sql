-- name: AddPartner :exec
INSERT INTO partners (user_id) VALUES ($1::uuid);

-- name: IsPartner :one
SELECT EXISTS(SELECT 1 FROM partners WHERE user_id = $1::uuid) AS is_partner;
