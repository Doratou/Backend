-- +goose Up
CREATE TYPE friendship_status AS ENUM ('PENDING', 'ACCEPTED', 'REJECTED');

CREATE TABLE friendships (
    id BIGSERIAL PRIMARY KEY,
    requester_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    receiver_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    status friendship_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_friendships_distinct_users CHECK (requester_user_id <> receiver_user_id)
);

CREATE UNIQUE INDEX idx_friendships_active_user_pair ON friendships (
    LEAST(requester_user_id, receiver_user_id),
    GREATEST(requester_user_id, receiver_user_id)
)
WHERE deleted_at IS NULL;

CREATE INDEX idx_friendships_requester_status ON friendships (requester_user_id, status, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_friendships_receiver_status ON friendships (receiver_user_id, status, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_friendships_deleted_at ON friendships (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS friendships;

DROP TYPE IF EXISTS friendship_status;
