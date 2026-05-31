CREATE TABLE IF NOT EXISTS "job_directions" (
                                  "id" serial PRIMARY KEY,
                                  "name" varchar NOT NULL,
                                  "tags" text[] NOT NULL DEFAULT '{Нижний Новгород, Офис, Полный день}',
                                  "description" varchar NOT NULL,
                                  "description_tags" text[] NOT NULL DEFAULT '{}'
);