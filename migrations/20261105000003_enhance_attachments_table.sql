-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- 1. Add SETTLEMENT_RECEIPT to attachment_type enum
ALTER TYPE attachment_type ADD VALUE IF NOT EXISTS 'SETTLEMENT_RECEIPT';

-- 2. Add media & auxiliary columns to attachments
ALTER TABLE attachments
    ADD COLUMN IF NOT EXISTS thumbnail_object_key TEXT,
    ADD COLUMN IF NOT EXISTS file_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS width INT,
    ADD COLUMN IF NOT EXISTS height INT,
    ADD COLUMN IF NOT EXISTS duration_seconds INT,
    ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. Add optimization indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_attachments_object_key ON attachments (object_key);
CREATE INDEX IF NOT EXISTS idx_attachments_file_hash ON attachments (file_hash) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by_user ON attachments (uploaded_by_user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attachments_group_id ON attachments (group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attachments_expense_id ON attachments (expense_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attachments_type ON attachments (type) WHERE deleted_at IS NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_attachments_type;
DROP INDEX IF EXISTS idx_attachments_expense_id;
DROP INDEX IF EXISTS idx_attachments_group_id;
DROP INDEX IF EXISTS idx_attachments_uploaded_by_user;
DROP INDEX IF EXISTS idx_attachments_file_hash;
DROP INDEX IF EXISTS idx_attachments_object_key;

ALTER TABLE attachments
    DROP COLUMN IF EXISTS is_public,
    DROP COLUMN IF EXISTS duration_seconds,
    DROP COLUMN IF EXISTS height,
    DROP COLUMN IF EXISTS width,
    DROP COLUMN IF EXISTS file_hash,
    DROP COLUMN IF EXISTS thumbnail_object_key;
