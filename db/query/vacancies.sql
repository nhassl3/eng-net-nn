-- name: GetVacancies :many
SELECT * FROM vacancies LIMIT $1 OFFSET $2;

-- name: GetVacancy :one
SELECT * FROM vacancies WHERE
                            (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
                            AND (sqlc.narg('name')::varchar IS NULL OR name=sqlc.narg('name')::varchar);

-- name: CreateVacancy :one
INSERT INTO vacancies (jd, name, description, required_exp, pay_day, skills) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateVacancy :one
UPDATE vacancies SET jd=$1, name=$2, description=$3, required_exp=$4, pay_day=$5, skills=$6, updated_at=now() RETURNING *;

-- name: RemoveVacancy :exec
DELETE FROM vacancies WHERE
                          (sqlc.narg('id')::uuid IS NULL OR id=sqlc.narg('id')::uuid)
                        AND(sqlc.narg('name')::varchar IS NULL OR name=sqlc.narg('name')::varchar);

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
