-- name: GetVacancies :many
SELECT * FROM vacancy_with_jd ORDER BY created_at DESC, id LIMIT $1 OFFSET $2;

-- name: GetVacancy :one
SELECT * FROM vacancy_with_jd WHERE
                            (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
                            AND (sqlc.narg('name')::varchar IS NULL OR name=sqlc.narg('name')::varchar)
                            AND (sqlc.narg('id')::uuid IS NOT NULL OR sqlc.narg('name')::varchar IS NOT NULL) LIMIT 1;

-- name: CreateVacancy :one
INSERT INTO vacancies (jd, name, description, required_exp, pay_day, skills) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateVacancy :one
UPDATE vacancies SET jd=$1, name=sqlc.arg('new_name')::varchar, description=$2, required_exp=$3, pay_day=$4, skills=$5, updated_at=now()
                 WHERE
                    (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
                    AND (sqlc.narg('name')::varchar IS NULL OR name=sqlc.narg('name')::varchar)
                    AND (sqlc.narg('id')::uuid IS NOT NULL OR sqlc.narg('name')::varchar IS NOT NULL)
                    RETURNING *;

-- name: RemoveVacancy :exec
DELETE FROM vacancies WHERE
                          (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
                        AND (sqlc.narg('name')::varchar IS NULL OR name=sqlc.narg('name')::varchar)
                        AND (sqlc.narg('id')::uuid IS NOT NULL OR sqlc.narg('name')::varchar IS NOT NULL);

-- name: GetJDs :many
SELECT * FROM job_directions ORDER BY id LIMIT $1 OFFSET $2;

-- name: GetJD :one
SELECT * FROM job_directions WHERE id=sqlc.arg('id')::bigint LIMIT 1;

-- name: CreateJobDirection :one
INSERT INTO job_directions (name, tags, description) VALUES (sqlc.arg('name')::varchar, sqlc.arg('tags')::text[], sqlc.arg('description')::text) RETURNING *;

-- name: UpdateJobDirection :one
UPDATE job_directions SET name=sqlc.narg('name')::varchar, tags=sqlc.narg('tags')::text[], description=sqlc.narg('description')::text
                      WHERE id=sqlc.arg('id')::bigint RETURNING *;

-- name: RemoveJobDirection :exec
DELETE FROM job_directions WHERE id=sqlc.arg('id')::bigint;

-- name: RespondToVacancy :one
INSERT INTO user_responds (full_name, phone_number, email, city, exp, description, resume, vacancy_id)
VALUES (
        sqlc.arg('full_name'),
        sqlc.narg('phone_number'),
        sqlc.arg('email'),
        sqlc.narg('city'),
        sqlc.arg('exp'),
        sqlc.narg('description'),
        sqlc.narg('resume'),
        sqlc.arg('vacancy_id')) RETURNING id;

-- name: GetRespondVacancies :many
SELECT * FROM user_responds ORDER BY created_at DESC, id LIMIT $1 OFFSET $2;

-- name: GetRespondVacancy :one
SELECT * FROM user_responds WHERE id=$1 LIMIT 1;
