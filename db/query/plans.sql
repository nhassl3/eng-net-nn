-- name: RequestPlan :one
INSERT INTO plans (full_name, direction, task_description, email) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetResponseFromRequest :one
SELECT * FROM link_user_with_plan WHERE plan_id=$1 LIMIT 1;

-- name: CreateLinkRequest :exec
INSERT INTO link_user_with_plan (user_id, plan_id) VALUES ($1, $2);

-- name: GetPlan :one
SELECT * FROM plans WHERE id=$1 LIMIT 1;

-- name: GetUserPlan :one
SELECT sqlc.embed(p), sqlc.embed(u)
FROM plans p
    INNER JOIN link_user_with_plan up ON p.id=up.plan_id
    INNER JOIN users u ON up.user_id=u.id
WHERE
    (sqlc.narg('user_id')::varchar IS NULL OR u.id=sqlc.narg('user_id')::varchar)
  AND p.id=sqlc.arg('plan_id');


-- name: GetDirection :one
SELECT name FROM directions WHERE id=$1 LIMIT 1;

-- name: GetAllPlans :many
SELECT * FROM plans LIMIT $1 OFFSET $2;
