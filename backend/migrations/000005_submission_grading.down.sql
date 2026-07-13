ALTER TABLE submissions
    DROP COLUMN IF EXISTS score,
    DROP COLUMN IF EXISTS feedback,
    DROP COLUMN IF EXISTS graded_at;
