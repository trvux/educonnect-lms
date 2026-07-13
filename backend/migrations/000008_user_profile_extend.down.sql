ALTER TABLE users
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS student_code,
    DROP COLUMN IF EXISTS avatar_path;
