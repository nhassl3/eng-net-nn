-- name: AddPartner :exec
INSERT INTO partners (user_id) VALUES ($1::uuid) ON CONFLICT DO NOTHING;
