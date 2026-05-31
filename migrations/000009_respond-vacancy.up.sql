CREATE TABLE IF NOT EXISTS user_responds (
    "id" uuid primary key default gen_random_uuid(),
    "full_name" varchar not null,
    "phone_number" varchar default '',
    "email" varchar unique not null,
    "city" varchar default '',
    "exp" varchar default 'нет',
    "description" text,
    "resume" varchar default '', /*link to minio where stored the resumes of users welcomed to job*/
    "vacancy_id" uuid not null,
    "created_at" timestamptz not null default now()
);

ALTER TABLE "user_responds" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id") DEFERRABLE INITIALLY IMMEDIATE;
