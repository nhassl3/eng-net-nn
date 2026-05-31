CREATE TABLE IF NOT EXISTS "plan" (
                        "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                        "full_name" varchar,
                        "direction" int NOT NULL,
                        "task_description" varchar,
                        "email" varchar UNIQUE NOT NULL,
                        "created_at" timestamptz NOT NULL DEFAULT (now())
);