-- +goose Up
ALTER TYPE group_status RENAME VALUE 'INACTIVE' TO 'SETTLING';
ALTER TYPE group_status ADD VALUE 'CLOSED';
ALTER TABLE groups ADD COLUMN closed_at TIMESTAMPTZ;

CREATE TYPE settlement_status AS ENUM ('PENDING', 'SUCCESS', 'REJECT', 'CANCELED');

CREATE TABLE settlements (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    from_member_id BIGINT NOT NULL,
    to_member_id BIGINT NOT NULL,
    amount_toman BIGINT NOT NULL CHECK (amount_toman > 0),
    status settlement_status NOT NULL DEFAULT 'PENDING',
    note TEXT,
    paid_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    rejection_reason TEXT,
    cancel_reason VARCHAR(40),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_settlements_id_group UNIQUE (id, group_id),
    CONSTRAINT fk_settlements_from_group FOREIGN KEY (from_member_id, group_id)
        REFERENCES group_members(id, group_id) ON DELETE RESTRICT,
    CONSTRAINT fk_settlements_to_group FOREIGN KEY (to_member_id, group_id)
        REFERENCES group_members(id, group_id) ON DELETE RESTRICT,
    CONSTRAINT chk_settlements_distinct_members CHECK (from_member_id <> to_member_id),
    CONSTRAINT chk_settlements_outcome CHECK (
        (status = 'PENDING' AND confirmed_at IS NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL)
        OR (status = 'SUCCESS' AND confirmed_at IS NOT NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'REJECT' AND rejected_at IS NOT NULL AND confirmed_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NOT NULL AND length(btrim(rejection_reason)) > 0 AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'CANCELED' AND canceled_at IS NOT NULL AND confirmed_at IS NULL AND rejected_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NOT NULL AND cancel_reason = 'RECIPIENT_LEFT' AND deleted_at IS NULL)
    )
);

CREATE INDEX idx_settlements_group_created ON settlements (group_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_settlements_pending_recipient ON settlements (to_member_id, id) WHERE status = 'PENDING' AND deleted_at IS NULL;
CREATE INDEX idx_settlements_pending_sender ON settlements (from_member_id, id) WHERE status = 'PENDING' AND deleted_at IS NULL;

ALTER TABLE attachments ADD COLUMN settlement_id BIGINT;
ALTER TABLE attachments ADD CONSTRAINT fk_attachments_settlement_group FOREIGN KEY (settlement_id, group_id)
    REFERENCES settlements(id, group_id) ON DELETE RESTRICT;
ALTER TABLE attachments ADD CONSTRAINT chk_attachments_settlement CHECK (
    settlement_id IS NULL OR (type = 'SETTLEMENT_RECEIPT' AND group_id IS NOT NULL AND expense_id IS NULL AND is_public = FALSE)
);
CREATE INDEX idx_attachments_settlement ON attachments (settlement_id, id) WHERE deleted_at IS NULL;
UPDATE attachments SET is_public = FALSE WHERE type = 'SETTLEMENT_RECEIPT';

-- +goose StatementBegin
CREATE FUNCTION cancel_settlements_on_recipient_departure() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status <> 'ACTIVE' OR NEW.left_at IS NOT NULL THEN
        UPDATE settlements SET status = 'CANCELED', canceled_at = CURRENT_TIMESTAMP,
            cancel_reason = 'RECIPIENT_LEFT', updated_at = CURRENT_TIMESTAMP
        WHERE to_member_id = NEW.id AND status = 'PENDING' AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
CREATE TRIGGER trg_cancel_settlements_on_recipient_departure
AFTER UPDATE OF status, left_at ON group_members
FOR EACH ROW EXECUTE FUNCTION cancel_settlements_on_recipient_departure();

-- +goose StatementBegin
CREATE FUNCTION guard_settlement_history() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'settlement history cannot be physically deleted' USING ERRCODE = '23514', CONSTRAINT = 'settlement_history_immutable';
    END IF;
    IF (NEW.group_id, NEW.from_member_id, NEW.to_member_id, NEW.amount_toman, NEW.note, NEW.paid_at, NEW.created_at)
        IS DISTINCT FROM (OLD.group_id, OLD.from_member_id, OLD.to_member_id, OLD.amount_toman, OLD.note, OLD.paid_at, OLD.created_at)
        OR OLD.status <> 'PENDING' OR OLD.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'settlement history is immutable' USING ERRCODE = '23514', CONSTRAINT = 'settlement_history_immutable';
    END IF;
    IF NEW.deleted_at IS NOT NULL AND NEW.status <> 'PENDING' THEN
        RAISE EXCEPTION 'only pending settlements can be deleted' USING ERRCODE = '23514', CONSTRAINT = 'settlement_history_immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
CREATE TRIGGER trg_guard_settlement_history BEFORE UPDATE OR DELETE ON settlements
FOR EACH ROW EXECUTE FUNCTION guard_settlement_history();

-- +goose StatementBegin
CREATE FUNCTION guard_group_financial_lifecycle() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL AND EXISTS (
        SELECT 1 FROM settlements WHERE group_id = OLD.id AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'group has settlement history' USING ERRCODE = '23514', CONSTRAINT = 'group_has_settlements';
    END IF;
    IF NEW.status IS DISTINCT FROM OLD.status THEN
        IF OLD.status::text = 'CLOSED' OR (OLD.status::text = 'SETTLING' AND NEW.status::text = 'ACTIVE')
            OR (OLD.status::text = 'ACTIVE' AND NEW.status::text = 'CLOSED') THEN
            RAISE EXCEPTION 'invalid group status transition' USING ERRCODE = '23514', CONSTRAINT = 'group_status_transition';
        END IF;
        IF NEW.status::text = 'CLOSED' AND EXISTS (
            SELECT 1 FROM settlements WHERE group_id = OLD.id AND status = 'PENDING' AND deleted_at IS NULL
        ) THEN
            RAISE EXCEPTION 'group has pending settlements' USING ERRCODE = '23514', CONSTRAINT = 'group_pending_settlements';
        END IF;
    END IF;
    IF OLD.status::text = 'CLOSED' AND NEW IS DISTINCT FROM OLD THEN
        RAISE EXCEPTION 'closed group is read-only' USING ERRCODE = '23514', CONSTRAINT = 'group_closed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
CREATE TRIGGER trg_guard_group_financial_lifecycle BEFORE UPDATE ON groups
FOR EACH ROW EXECUTE FUNCTION guard_group_financial_lifecycle();

-- +goose Down
DROP TRIGGER trg_guard_group_financial_lifecycle ON groups;
DROP FUNCTION guard_group_financial_lifecycle();
DROP TRIGGER trg_cancel_settlements_on_recipient_departure ON group_members;
DROP FUNCTION cancel_settlements_on_recipient_departure();
DROP TRIGGER trg_guard_settlement_history ON settlements;
DROP FUNCTION guard_settlement_history();
DROP INDEX idx_attachments_settlement;
ALTER TABLE attachments DROP CONSTRAINT chk_attachments_settlement;
ALTER TABLE attachments DROP CONSTRAINT fk_attachments_settlement_group;
ALTER TABLE attachments DROP COLUMN settlement_id;
DROP TABLE settlements;
DROP TYPE settlement_status;
ALTER TABLE groups DROP COLUMN closed_at;
CREATE TYPE group_status_legacy AS ENUM ('ACTIVE', 'INACTIVE');
ALTER TABLE groups ALTER COLUMN status DROP DEFAULT;
ALTER TABLE groups ALTER COLUMN status TYPE group_status_legacy
    USING (CASE WHEN status::text = 'ACTIVE' THEN 'ACTIVE' ELSE 'INACTIVE' END)::group_status_legacy;
DROP TYPE group_status;
ALTER TYPE group_status_legacy RENAME TO group_status;
ALTER TABLE groups ALTER COLUMN status SET DEFAULT 'ACTIVE';
