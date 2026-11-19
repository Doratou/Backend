-- +goose Up
ALTER TABLE settlements
    DROP CONSTRAINT chk_settlements_outcome,
    ADD CONSTRAINT chk_settlements_outcome CHECK (
        (status = 'PENDING' AND confirmed_at IS NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL)
        OR (status = 'SUCCESS' AND confirmed_at IS NOT NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'REJECT' AND rejected_at IS NOT NULL AND confirmed_at IS NULL AND canceled_at IS NULL AND (rejection_reason IS NULL OR length(btrim(rejection_reason)) > 0) AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'CANCELED' AND canceled_at IS NOT NULL AND confirmed_at IS NULL AND rejected_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NOT NULL AND cancel_reason = 'RECIPIENT_LEFT' AND deleted_at IS NULL)
    );

-- +goose Down
ALTER TABLE settlements DISABLE TRIGGER trg_guard_settlement_history;
UPDATE settlements
SET rejection_reason = 'No rejection reason was recorded'
WHERE status = 'REJECT' AND rejection_reason IS NULL;
ALTER TABLE settlements ENABLE TRIGGER trg_guard_settlement_history;

ALTER TABLE settlements
    DROP CONSTRAINT chk_settlements_outcome,
    ADD CONSTRAINT chk_settlements_outcome CHECK (
        (status = 'PENDING' AND confirmed_at IS NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL)
        OR (status = 'SUCCESS' AND confirmed_at IS NOT NULL AND rejected_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'REJECT' AND rejected_at IS NOT NULL AND confirmed_at IS NULL AND canceled_at IS NULL AND rejection_reason IS NOT NULL AND length(btrim(rejection_reason)) > 0 AND cancel_reason IS NULL AND deleted_at IS NULL)
        OR (status = 'CANCELED' AND canceled_at IS NOT NULL AND confirmed_at IS NULL AND rejected_at IS NULL AND rejection_reason IS NULL AND cancel_reason IS NOT NULL AND cancel_reason = 'RECIPIENT_LEFT' AND deleted_at IS NULL)
    );
