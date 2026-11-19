-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cancel_settlements_on_recipient_departure() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status <> 'ACTIVE' OR NEW.left_at IS NOT NULL THEN
        UPDATE settlements SET status = 'CANCELED', canceled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
        WHERE to_member_id = NEW.id AND status = 'PENDING' AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

ALTER TABLE settlements
    DROP CONSTRAINT chk_settlements_outcome,
    DROP COLUMN cancel_reason,
    ADD CONSTRAINT chk_settlements_outcome CHECK (
        (status = 'PENDING' AND confirmed_at IS NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL)
        OR (status = 'SUCCESS' AND confirmed_at IS NOT NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'REJECT' AND rejected_at IS NOT NULL AND confirmed_at IS NULL AND canceled_at IS NULL AND (rejection_reason IS NULL OR length(btrim(rejection_reason)) > 0) AND deleted_at IS NULL)
        OR (status = 'CANCELED' AND canceled_at IS NOT NULL AND confirmed_at IS NULL AND rejected_at IS NULL AND rejection_reason IS NULL AND deleted_at IS NULL)
    );

-- +goose Down
ALTER TABLE settlements ADD COLUMN cancel_reason VARCHAR(40);
ALTER TABLE settlements DISABLE TRIGGER trg_guard_settlement_history;
UPDATE settlements SET cancel_reason = 'RECIPIENT_LEFT' WHERE status = 'CANCELED';
ALTER TABLE settlements ENABLE TRIGGER trg_guard_settlement_history;

ALTER TABLE settlements
    DROP CONSTRAINT chk_settlements_outcome,
    ADD CONSTRAINT chk_settlements_outcome CHECK (
        (status = 'PENDING' AND confirmed_at IS NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL)
        OR (status = 'SUCCESS' AND confirmed_at IS NOT NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'REJECT' AND rejected_at IS NOT NULL AND confirmed_at IS NULL AND canceled_at IS NULL AND (rejection_reason IS NULL OR length(btrim(rejection_reason)) > 0) AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'CANCELED' AND canceled_at IS NOT NULL AND confirmed_at IS NULL AND rejected_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NOT NULL AND cancel_reason = 'RECIPIENT_LEFT' AND deleted_at IS NULL)
    );

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cancel_settlements_on_recipient_departure() RETURNS TRIGGER AS $$
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
