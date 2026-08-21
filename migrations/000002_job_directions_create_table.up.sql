CREATE TABLE IF NOT EXISTS "job_directions" (
                                  "id" serial PRIMARY KEY,
                                  "name" varchar NOT NULL,
                                  "tags" text[] NOT NULL DEFAULT '{}',
                                  "description" varchar NOT NULL
);