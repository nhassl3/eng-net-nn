CREATE TABLE IF NOT EXISTS cases_stats (
                               id          BIGSERIAL PRIMARY KEY,
                               case_id  BIGSERIAL NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
                               value       VARCHAR(100) NOT NULL,
                               label       VARCHAR(100) NOT NULL
);