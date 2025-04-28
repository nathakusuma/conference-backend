CREATE TABLE sessions
(
    token   CHAR(32) PRIMARY KEY,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
CREATE UNIQUE INDEX sessions_user_id_key ON sessions (user_id);
