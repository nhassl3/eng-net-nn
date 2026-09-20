CREATE TABLE IF NOT EXISTS case_stats (
                               id          BIGSERIAL PRIMARY KEY,
                               case_id  BIGINT NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
                               value       VARCHAR(100) NOT NULL,
                               label       VARCHAR(100) NOT NULL
);