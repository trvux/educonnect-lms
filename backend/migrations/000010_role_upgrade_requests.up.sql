CREATE TABLE IF NOT EXISTS role_upgrade_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    reviewed_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_role_upgrade_pending_unique ON role_upgrade_requests(user_id) WHERE status = 'pending';
