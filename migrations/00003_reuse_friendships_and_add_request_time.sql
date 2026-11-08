-- +goose Up
LOCK TABLE friendships IN ACCESS EXCLUSIVE MODE;

ALTER TABLE friendships ADD COLUMN requested_at TIMESTAMPTZ;

UPDATE friendships SET requested_at = created_at;

ALTER TABLE friendships ALTER COLUMN requested_at SET NOT NULL;
ALTER TABLE friendships ALTER COLUMN requested_at SET DEFAULT CURRENT_TIMESTAMP;

WITH ranked_friendships AS (
    SELECT id, ROW_NUMBER() OVER (
        PARTITION BY LEAST(requester_user_id, receiver_user_id), GREATEST(requester_user_id, receiver_user_id)
        ORDER BY created_at DESC, id DESC
    ) AS position
    FROM friendships
)
DELETE FROM friendships
USING ranked_friendships
WHERE friendships.id = ranked_friendships.id AND ranked_friendships.position > 1;

DROP INDEX idx_friendships_active_user_pair;

CREATE UNIQUE INDEX idx_friendships_user_pair ON friendships (
    LEAST(requester_user_id, receiver_user_id),
    GREATEST(requester_user_id, receiver_user_id)
);

DROP INDEX idx_friendships_requester_status;
DROP INDEX idx_friendships_receiver_status;

CREATE INDEX idx_friendships_requester_status ON friendships (requester_user_id, status, requested_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_friendships_receiver_status ON friendships (receiver_user_id, status, requested_at DESC, id DESC)
WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX idx_friendships_user_pair;

CREATE UNIQUE INDEX idx_friendships_active_user_pair ON friendships (
    LEAST(requester_user_id, receiver_user_id),
    GREATEST(requester_user_id, receiver_user_id)
)
WHERE deleted_at IS NULL;

DROP INDEX idx_friendships_requester_status;
DROP INDEX idx_friendships_receiver_status;

CREATE INDEX idx_friendships_requester_status ON friendships (requester_user_id, status, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_friendships_receiver_status ON friendships (receiver_user_id, status, id DESC)
WHERE deleted_at IS NULL;

UPDATE friendships SET created_at = requested_at;

ALTER TABLE friendships DROP COLUMN requested_at;
