-- +goose Up
ALTER TABLE polls ADD COLUMN allow_multiple_choices BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE poll_votes DROP CONSTRAINT uq_poll_votes_poll_member;
ALTER TABLE poll_votes ADD CONSTRAINT uq_poll_votes_poll_member_option UNIQUE (poll_id, member_id, option_id);

-- +goose Down
LOCK TABLE polls, poll_votes IN ACCESS EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM poll_votes
        GROUP BY poll_id, member_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'Cannot roll back poll multiple choices while a member has multiple stored option rows; rolling back would lose vote data';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE poll_votes DROP CONSTRAINT uq_poll_votes_poll_member_option;
ALTER TABLE poll_votes ADD CONSTRAINT uq_poll_votes_poll_member UNIQUE (poll_id, member_id);

ALTER TABLE polls DROP COLUMN allow_multiple_choices;
