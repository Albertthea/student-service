CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    grade INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
