CREATE TABLE logs (
    id BIGINT PRIMARY KEY,
    unix_ts BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    event_name SMALLINT NOT NULL
);