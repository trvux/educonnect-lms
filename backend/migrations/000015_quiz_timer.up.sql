ALTER TABLE assignments ADD COLUMN IF NOT EXISTS time_limit_minutes INTEGER;

CREATE TABLE IF NOT EXISTS quiz_attempts (
    id SERIAL PRIMARY KEY,
    assignment_id INTEGER NOT NULL REFERENCES assignments(id),
    student_id INTEGER NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (assignment_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_assignment ON quiz_attempts(assignment_id);
