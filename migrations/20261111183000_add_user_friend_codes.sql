-- +goose Up
ALTER TABLE users ADD COLUMN friend_code VARCHAR(15);

CREATE UNIQUE INDEX idx_users_friend_code ON users (friend_code);

-- +goose StatementBegin
CREATE FUNCTION generate_user_friend_code() RETURNS TEXT
LANGUAGE plpgsql VOLATILE AS $$
DECLARE
    alphabet CONSTANT TEXT := '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
    candidate TEXT;
    attempt INTEGER;
    position INTEGER;
BEGIN
    FOR attempt IN 1..100 LOOP
        candidate := 'DR-';
        FOR position IN 1..12 LOOP
            candidate := candidate || SUBSTR(alphabet, (GET_BYTE(UUID_SEND(GEN_RANDOM_UUID()), 0) & 31) + 1, 1);
        END LOOP;
        PERFORM PG_ADVISORY_XACT_LOCK(HASHTEXTEXTENDED(candidate, 0));
        IF NOT EXISTS (SELECT 1 FROM users WHERE friend_code = candidate) THEN
            RETURN candidate;
        END IF;
    END LOOP;
    RAISE EXCEPTION 'could not generate a unique friend code';
END;
$$;
-- +goose StatementEnd

UPDATE users SET friend_code = generate_user_friend_code() WHERE friend_code IS NULL;

ALTER TABLE users
    ALTER COLUMN friend_code SET DEFAULT generate_user_friend_code(),
    ALTER COLUMN friend_code SET NOT NULL,
    ADD CONSTRAINT chk_users_friend_code CHECK (friend_code ~ '^DR-[23456789A-HJ-NP-Z]{12}$');

-- +goose Down
ALTER TABLE users DROP COLUMN friend_code;

DROP FUNCTION generate_user_friend_code();
