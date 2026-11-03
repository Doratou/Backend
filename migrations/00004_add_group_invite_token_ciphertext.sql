-- +goose Up
ALTER TABLE group_invites ADD COLUMN token_ciphertext TEXT;

-- +goose Down
ALTER TABLE group_invites DROP COLUMN IF EXISTS token_ciphertext;
