-- +goose Up
CREATE TYPE notification_event AS ENUM (
    'friend_request.received',
    'friend_request.responded',
    'settlement.paid',
    'poll.created',
    'poll.closed',
    'draw.created',
    'draw.result',
    'checklist.updated',
    'admin_report.answered',
    'group.member_joined',
    'group.member_left',
    'token.price_updated',
    'group.announcement'
);

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    recipient_id BIGINT NOT NULL,
    actor_id BIGINT,
    group_id BIGINT,
    event notification_event NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_notifications_recipient_timeline 
ON notifications (recipient_id, id DESC) 
WHERE deleted_at IS NULL;

CREATE TABLE notification_reads (
    user_id BIGINT PRIMARY KEY,
    last_read_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS notification_reads;
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_event;
