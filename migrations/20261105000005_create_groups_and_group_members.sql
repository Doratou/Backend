-- +goose Up
CREATE TYPE group_status AS ENUM ('ACTIVE', 'INACTIVE');

CREATE TABLE groups (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    cover_attachment_id BIGINT REFERENCES attachments(id) ON UPDATE CASCADE ON DELETE SET NULL,
    status group_status NOT NULL DEFAULT 'ACTIVE',
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    archived_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_members (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status VARCHAR(15) NOT NULL DEFAULT 'ACTIVE',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMPTZ,
    CONSTRAINT uq_group_members_group_user UNIQUE (group_id, user_id)
);

CREATE INDEX idx_groups_created_by_user_id ON groups(created_by_user_id);
CREATE INDEX idx_groups_deleted_at ON groups(deleted_at);
CREATE INDEX idx_group_members_user_status ON group_members(user_id, status);

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TYPE IF EXISTS group_status;
