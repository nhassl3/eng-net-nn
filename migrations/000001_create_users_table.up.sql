CREATE TABLE IF NOT EXISTS "users" (
                         "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                         "username" varchar,
                         "full_name" varchar,
                         "email" varchar UNIQUE NOT NULL,
                         "created_at" timestamptz NOT NULL DEFAULT (now()),
                         "updated_at" timestamptz NOT NULL DEFAULT (now())
);