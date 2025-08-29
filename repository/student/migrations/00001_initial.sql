-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    grade      INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_students_grade ON students (grade);

-- +goose Down
DROP INDEX IF EXISTS idx_students_grade;
DROP TABLE IF EXISTS students;
