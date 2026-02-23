BEGIN;

CREATE TABLE submissions (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255),
    timestamp TIMESTAMP,
    submission_count INT
);

COMMIT;
