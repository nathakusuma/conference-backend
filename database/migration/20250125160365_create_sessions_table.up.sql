CREATE TABLE sessions
(
    user_id    UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    token      CHAR(64)  NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
