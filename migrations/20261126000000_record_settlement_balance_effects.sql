-- +goose Up
ALTER TABLE settlements
    ADD COLUMN IF NOT EXISTS balance_legs JSONB,
    ADD COLUMN IF NOT EXISTS balance_revision CHAR(64);

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'settlements'::regclass AND conname = 'chk_settlement_balance_legs'
    ) THEN
        ALTER TABLE settlements ADD CONSTRAINT chk_settlement_balance_legs CHECK (
            balance_legs IS NULL OR (status = 'SUCCESS' AND jsonb_typeof(balance_legs) = 'array')
        );
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS financial_report_opening_states (
    group_id BIGINT PRIMARY KEY REFERENCES groups(id) ON DELETE CASCADE,
    state JSONB NOT NULL
);

INSERT INTO financial_report_opening_states (group_id, state)
SELECT legacy.group_id, jsonb_build_object(
    'members', COALESCE((
        SELECT jsonb_agg(jsonb_build_object(
            'id', m.id, 'joined_at', m.joined_at, 'left_at', m.left_at,
            'active', m.status = 'ACTIVE' AND m.left_at IS NULL
        ) ORDER BY m.id)
        FROM group_members m WHERE m.group_id = legacy.group_id
    ), '[]'::jsonb),
    'charges', COALESCE((
        SELECT jsonb_agg(jsonb_build_object(
            'from_member_id', sh.member_id, 'to_member_id', e.created_by_member_id,
            'amount_toman', sh.share_amount_toman,
            'recorded_at', GREATEST(e.created_at, sh.created_at)
        ) ORDER BY GREATEST(e.created_at, sh.created_at), sh.id)
        FROM expenses e JOIN expense_shares sh ON sh.expense_id = e.id
        WHERE e.group_id = legacy.group_id AND e.deleted_at IS NULL AND sh.deleted_at IS NULL
          AND sh.member_id <> e.created_by_member_id AND sh.share_amount_toman > 0
    ), '[]'::jsonb)
)
FROM (SELECT DISTINCT group_id FROM settlements WHERE status = 'SUCCESS' AND deleted_at IS NULL AND balance_legs IS NULL) legacy
ON CONFLICT (group_id) DO NOTHING;

-- +goose Down
DROP TABLE financial_report_opening_states;
ALTER TABLE settlements DROP CONSTRAINT chk_settlement_balance_legs, DROP COLUMN balance_legs, DROP COLUMN balance_revision;
