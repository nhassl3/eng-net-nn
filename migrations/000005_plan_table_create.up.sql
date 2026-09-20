CREATE TABLE IF NOT EXISTS "plans" (
                        "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                        "full_name" varchar,
                        "direction" int NOT NULL,
                        "task_description" varchar,
                        "email" varchar NOT NULL,
                        "active" bool NOT NULL DEFAULT TRUE ,
                        "created_at" timestamptz NOT NULL DEFAULT (now()),
                        "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_plans_one_active_per_email ON plans(email) WHERE active=TRUE;