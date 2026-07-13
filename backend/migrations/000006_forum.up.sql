CREATE TABLE IF NOT EXISTS forum_posts (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id),
    author_id INTEGER NOT NULL REFERENCES users(id),
    parent_id INTEGER REFERENCES forum_posts(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_forum_posts_course ON forum_posts(course_id);
CREATE INDEX IF NOT EXISTS idx_forum_posts_parent ON forum_posts(parent_id);
