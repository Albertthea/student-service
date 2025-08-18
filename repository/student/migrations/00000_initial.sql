CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    grade INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    middle_name TEXT,
    status TEXT,
    home_address JSONB,
    course_grades JSONB,
    friends TEXT[],
    local JSONB,
    exchange JSONB
);

CREATE INDEX IF NOT EXISTS idx_students_grade ON students (grade);