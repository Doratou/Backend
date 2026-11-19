-- +goose Up
CREATE TYPE poll_status AS ENUM ('ACTIVE', 'INACTIVE');
CREATE TYPE poll_visibility AS ENUM ('PUBLIC', 'PRIVATE');

CREATE TABLE polls (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_by_member_id BIGINT NOT NULL REFERENCES group_members(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    question VARCHAR(500) NOT NULL,
    visibility poll_visibility NOT NULL DEFAULT 'PRIVATE',
    status poll_status NOT NULL DEFAULT 'ACTIVE',
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_polls_question_not_empty CHECK (BTRIM(question) <> ''),
    CONSTRAINT chk_polls_closed_at CHECK (
        (status = 'ACTIVE' AND closed_at IS NULL)
        OR (status = 'INACTIVE' AND closed_at IS NOT NULL)
    )
);

CREATE TABLE poll_options (
    id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL REFERENCES polls(id) ON UPDATE CASCADE ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_poll_options_title_not_empty CHECK (BTRIM(title) <> ''),
    CONSTRAINT uq_poll_options_poll_id_id UNIQUE (poll_id, id)
);

CREATE TABLE poll_votes (
    id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL REFERENCES polls(id) ON UPDATE CASCADE ON DELETE CASCADE,
    member_id BIGINT NOT NULL REFERENCES group_members(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    option_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_poll_votes_poll_option FOREIGN KEY (poll_id, option_id)
        REFERENCES poll_options(poll_id, id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT uq_poll_votes_poll_member UNIQUE (poll_id, member_id)
);

CREATE INDEX idx_polls_group_status ON polls (group_id, status, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_polls_created_by_member_id ON polls (created_by_member_id);
CREATE INDEX idx_polls_deleted_at ON polls (deleted_at);

CREATE UNIQUE INDEX idx_poll_options_active_title ON poll_options (poll_id, LOWER(BTRIM(title)))
WHERE deleted_at IS NULL;

CREATE INDEX idx_poll_options_deleted_at ON poll_options (deleted_at);

CREATE INDEX idx_poll_votes_poll_option ON poll_votes (poll_id, option_id, id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_poll_votes_member_id ON poll_votes (member_id);
CREATE INDEX idx_poll_votes_deleted_at ON poll_votes (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS polls;

DROP TYPE IF EXISTS poll_visibility;
DROP TYPE IF EXISTS poll_status;
