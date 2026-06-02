-- name: RequestPlan :one
INSERT INTO plan (full_name, direction, task_description, email) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetResponseFromRequest :one
SELECT * FROM link_user_with_plan WHERE plan_id=$1;

-- name: CreateLinkRequest :exec
INSERT INTO link_user_with_plan (user_id, plan_id) VALUES ($1, $2);

-- name: GetPlan :one
SELECT * FROM plan WHERE id=$1;

-- name: GetDirection :one
SELECT name FROM directions WHERE id=$1;
