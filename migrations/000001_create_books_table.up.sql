CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE books (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    author      VARCHAR(255) NOT NULL,
    description TEXT,
    year        INTEGER,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_books_deleted_at ON books (deleted_at);
