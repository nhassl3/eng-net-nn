CREATE TABLE IF NOT EXISTS "vacancies" (
                             "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                             "jd" int NOT NULL,
                             "name" varchar,
                             "description" varchar,
                             "required_exp" varchar default '',
                             "pay_day" numeric(10,2) NOT NULL CHECK (pay_day > 0),
                             "skills" text[] NOT NULL DEFAULT '{}',
                             "created_at" timestamptz NOT NULL DEFAULT (now()),
                             "updated_at" timestamptz NOT NULL DEFAULT (now()),
                             CONSTRAINT jd_name_vacancy_unique UNIQUE (jd, name)
);