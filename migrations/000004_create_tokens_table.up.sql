CREATE TABLE tokens (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id),
    token_hash VARCHAR(64) NOT NULL,
    type       VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tokens_token_hash ON tokens (token_hash);
CREATE INDEX idx_tokens_user_id    ON tokens (user_id);
