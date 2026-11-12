-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TYPE expense_type AS ENUM ('NORMAL', 'PREMIUM');

CREATE TABLE expenses (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    total_amount_toman BIGINT NOT NULL DEFAULT 0,
    created_by_member_id BIGINT NOT NULL REFERENCES group_members(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    expense_type expense_type NOT NULL DEFAULT 'NORMAL',
    occurred_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_expenses_group_id ON expenses (group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_created_by_member ON expenses (created_by_member_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_deleted_at ON expenses (deleted_at);

-- Add foreign keys to attachments
ALTER TABLE attachments
    ADD CONSTRAINT fk_attachments_group
    FOREIGN KEY (group_id) REFERENCES groups(id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    ADD CONSTRAINT fk_attachments_expense
    FOREIGN KEY (expense_id) REFERENCES expenses(id)
    ON UPDATE CASCADE ON DELETE SET NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE attachments
    DROP CONSTRAINT IF EXISTS fk_attachments_expense,
    DROP CONSTRAINT IF EXISTS fk_attachments_group;

DROP TABLE IF EXISTS expenses CASCADE;

DROP TYPE IF EXISTS expense_type;
