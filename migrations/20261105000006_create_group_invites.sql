-- +goose Up
CREATE TABLE group_invites (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_by_member_id BIGINT NOT NULL REFERENCES group_members(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    token_hash CHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    max_uses INTEGER CHECK (max_uses IS NULL OR max_uses > 0),
    used_count INTEGER NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_invites_group_id ON group_invites(group_id);

-- +goose Down
DROP TABLE IF EXISTS group_invites;
