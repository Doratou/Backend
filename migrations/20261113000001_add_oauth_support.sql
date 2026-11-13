-- +goose Up
-- 1. Make password_hash nullable — OAuth users have no password
ALTER TABLE
    users
ALTER COLUMN
    password_hash DROP NOT NULL,
ALTER COLUMN
    password_hash
SET
    DEFAULT NULL;

-- 2. OAuth identity table (Option A: separate table, provider-agnostic)
CREATE TABLE user_oauth_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    provider_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_oauth_provider_id UNIQUE (provider, provider_id)
);

CREATE INDEX idx_oauth_user_id ON user_oauth_accounts (user_id);

-- +goose Down
ALTER TABLE
    users
ALTER COLUMN
    password_hash
SET
    NOT NULL,
ALTER COLUMN
    password_hash DROP DEFAULT;

DROP TABLE IF EXISTS user_oauth_accounts;