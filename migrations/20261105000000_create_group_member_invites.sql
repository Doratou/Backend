-- +goose Up
CREATE TYPE group_invite_status AS ENUM ('PENDING', 'ACCEPTED', 'REJECTED');

CREATE TABLE group_member_invites (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
    sender_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    recipient_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status group_invite_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMPTZ,
    CONSTRAINT chk_group_member_invites_distinct_users CHECK (sender_user_id <> recipient_user_id)
);

CREATE INDEX idx_group_member_invites_recipient_status ON group_member_invites(recipient_user_id, status, created_at DESC);
CREATE INDEX idx_group_member_invites_group_created ON group_member_invites(group_id, created_at DESC);
CREATE UNIQUE INDEX uq_group_member_invites_pending ON group_member_invites(group_id, recipient_user_id) WHERE status = 'PENDING';

-- +goose Down
DROP TABLE IF EXISTS group_member_invites;
DROP TYPE IF EXISTS group_invite_status;
