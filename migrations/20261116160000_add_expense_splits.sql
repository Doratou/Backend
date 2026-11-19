-- +goose Up
CREATE TYPE expense_split_type AS ENUM ('EQUAL', 'EXACT', 'PERCENTAGE', 'WEIGHTED');

ALTER TABLE expenses ADD COLUMN split_type expense_split_type NOT NULL DEFAULT 'EQUAL';
ALTER TABLE expenses ADD CONSTRAINT uq_expenses_id_group UNIQUE (id, group_id);
ALTER TABLE group_members ADD CONSTRAINT uq_group_members_id_group UNIQUE (id, group_id);

CREATE TABLE expense_shares (
    id BIGSERIAL PRIMARY KEY,
    expense_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,
    share_amount_toman BIGINT NOT NULL CHECK (share_amount_toman >= 0),
    split_value BIGINT NOT NULL CHECK (split_value >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_expense_shares_expense_member UNIQUE (expense_id, member_id),
    CONSTRAINT fk_expense_shares_expense_group FOREIGN KEY (expense_id, group_id)
        REFERENCES expenses(id, group_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_expense_shares_member_group FOREIGN KEY (member_id, group_id)
        REFERENCES group_members(id, group_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_expense_shares_member ON expense_shares (member_id, expense_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_group_occurred_at ON expenses (group_id, occurred_at DESC, id DESC) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX idx_expenses_group_occurred_at;
DROP TABLE expense_shares;
ALTER TABLE group_members DROP CONSTRAINT uq_group_members_id_group;
ALTER TABLE expenses DROP CONSTRAINT uq_expenses_id_group;
ALTER TABLE expenses DROP COLUMN split_type;
DROP TYPE expense_split_type;
