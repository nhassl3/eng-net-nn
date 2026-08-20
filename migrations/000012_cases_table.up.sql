CREATE TABLE IF NOT EXISTS cases (
    id BIGSERIAL primary key,
    title VARCHAR(120) not null,
    description TEXT not null,
    label VARCHAR(50) default '',
    photo VARCHAR(120) default ''
);
