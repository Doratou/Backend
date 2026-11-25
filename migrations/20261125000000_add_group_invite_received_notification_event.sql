-- +goose Up
ALTER TYPE notification_event ADD VALUE IF NOT EXISTS 'group.invite_received';

-- +goose Down
-- PostgreSQL does not support dropping a single ENUM value without recreating the type.
