-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION profile_completion_steps(account users)
RETURNS TABLE (full_name BOOLEAN, profile_photo BOOLEAN, verified_email BOOLEAN,
    verified_phone BOOLEAN, telegram_connected BOOLEAN, reward_tokens BIGINT)
LANGUAGE sql STABLE AS $$
    SELECT
        (account).name ~ '^[[:space:]]*[^[:space:]]+[[:space:]]+[^[:space:]]+',
        EXISTS (
            SELECT 1 FROM attachments AS photo
            WHERE photo.id = (account).avatar_attachment_id
              AND photo.uploaded_by_user_id = (account).id
              AND photo.type = 'PROFILE_IMAGE'
              AND photo.deleted_at IS NULL
        ),
        (account).is_email_verified AND COALESCE(btrim((account).email) <> '', FALSE),
        (account).is_phone_verified AND COALESCE(btrim((account).phone) <> '', FALSE),
        COALESCE((account).telegram_user_id > 0, FALSE),
        3::BIGINT;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION award_profile_completion_tokens() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
DECLARE
    steps RECORD;
BEGIN
    IF EXISTS (
        SELECT 1 FROM wallet_transactions
        WHERE user_id = NEW.id AND idempotency_key = 'profile completion reward'
    ) THEN
        RETURN NEW;
    END IF;

    SELECT * INTO steps FROM profile_completion_steps(NEW);
    IF NOT (steps.full_name AND steps.profile_photo AND steps.verified_email
        AND steps.verified_phone AND steps.telegram_connected) THEN
        RETURN NEW;
    END IF;

    INSERT INTO wallet_transactions (
        user_id, type, status, description, token_amount, balance_after, idempotency_key
    ) VALUES (
        NEW.id, 'REWARD', 'SUCCEEDED', 'Profile completion reward', steps.reward_tokens,
        NEW.wallet + steps.reward_tokens, 'profile completion reward'
    ) ON CONFLICT (user_id, idempotency_key) DO NOTHING;

    IF FOUND THEN
        NEW.wallet := NEW.wallet + steps.reward_tokens;
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_award_profile_completion_tokens
BEFORE UPDATE OF name, avatar_attachment_id, email, is_email_verified, phone,
    is_phone_verified, telegram_user_id, status, deleted_at ON users
FOR EACH ROW
WHEN (NEW.status = 'ACTIVE' AND NEW.deleted_at IS NULL AND NEW.telegram_user_id > 0)
EXECUTE FUNCTION award_profile_completion_tokens();

UPDATE users AS account SET name = account.name
WHERE account.status = 'ACTIVE' AND account.deleted_at IS NULL AND account.telegram_user_id > 0
  AND EXISTS (
    SELECT 1 FROM profile_completion_steps(account) AS steps
    WHERE steps.full_name AND steps.profile_photo AND steps.verified_email
      AND steps.verified_phone AND steps.telegram_connected
  );

-- +goose Down
DROP TRIGGER trg_award_profile_completion_tokens ON users;
DROP FUNCTION award_profile_completion_tokens();
DROP FUNCTION profile_completion_steps(users);
