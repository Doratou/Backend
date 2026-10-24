-- +goose Up
-- SQL in this section is executed when the migration is applied.

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('USER', 'ADMIN');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
        CREATE TYPE user_status AS ENUM ('ACTIVE', 'INACTIVE');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attachment_type') THEN
        CREATE TYPE attachment_type AS ENUM ('PROFILE_IMAGE', 'GROUP_COVER', 'GALLERY_IMAGE', 'GALLERY_VIDEO', 'EXPENSE_RECEIPT');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY,
    uploaded_by_user_id BIGINT,
    group_id BIGINT,
    expense_id BIGINT,
    type attachment_type NOT NULL,
    original_filename VARCHAR(255),
    object_key TEXT,
    mime_type VARCHAR(100),
    file_size_bytes BIGINT,
    uploaded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(16),
    password_hash TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_attachment_id BIGINT REFERENCES attachments(id) ON UPDATE CASCADE ON DELETE SET NULL,
    system_role user_role NOT NULL DEFAULT 'USER',
    status user_status NOT NULL DEFAULT 'ACTIVE',
    wallet BIGINT NOT NULL DEFAULT 0,
    telegram_chat_id BIGINT,
    telegram_user_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Foreign key for uploaded_by_user_id in attachments table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_attachments_uploaded_by_user'
    ) THEN
        ALTER TABLE attachments
            ADD CONSTRAINT fk_attachments_uploaded_by_user
            FOREIGN KEY (uploaded_by_user_id) REFERENCES users(id)
            ON UPDATE CASCADE ON DELETE SET NULL;
    END IF;
END$$;

-- Partial unique indexes on email and phone (only for non-deleted, non-null rows)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email)
WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone)
WHERE deleted_at IS NULL AND phone IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_attachments_deleted_at ON attachments (deleted_at);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE attachments DROP CONSTRAINT IF EXISTS fk_attachments_uploaded_by_user;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS attachments CASCADE;
DROP TYPE IF EXISTS attachment_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
