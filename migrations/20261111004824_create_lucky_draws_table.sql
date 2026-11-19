-- +goose Up
CREATE TYPE lucky_draw_status AS ENUM ('DRAFT', 'COMPLETED');

-- +goose StatementBegin
CREATE FUNCTION valid_lucky_draw_candidates(value JSONB) RETURNS BOOLEAN
LANGUAGE plpgsql IMMUTABLE AS $$
DECLARE
    candidate JSONB;
    candidate_title TEXT;
    candidate_weight NUMERIC;
    seen_titles TEXT[] := ARRAY[]::TEXT[];
BEGIN
    IF value IS NULL OR jsonb_typeof(value) <> 'array' THEN
        RETURN FALSE;
    END IF;
    IF jsonb_array_length(value) < 2 OR jsonb_array_length(value) > 100 THEN
        RETURN FALSE;
    END IF;
    FOR candidate IN SELECT * FROM jsonb_array_elements(value) LOOP
        IF jsonb_typeof(candidate) <> 'object'
            OR jsonb_typeof(candidate -> 'title') IS DISTINCT FROM 'string'
            OR jsonb_typeof(candidate -> 'weight') IS DISTINCT FROM 'number' THEN
            RETURN FALSE;
        END IF;
        IF (candidate ->> 'weight') !~ '^[1-9][0-9]*$' THEN
            RETURN FALSE;
        END IF;
        candidate_title := BTRIM(candidate ->> 'title');
        candidate_weight := (candidate ->> 'weight')::NUMERIC;
        IF candidate_title = '' OR CHAR_LENGTH(candidate ->> 'title') > 255
            OR candidate_weight < 1 OR candidate_weight > 1000000
            OR candidate_weight <> TRUNC(candidate_weight) THEN
            RETURN FALSE;
        END IF;
        candidate_title := LOWER(candidate_title);
        IF candidate_title = ANY(seen_titles) THEN
            RETURN FALSE;
        END IF;
        seen_titles := ARRAY_APPEND(seen_titles, candidate_title);
    END LOOP;
    RETURN TRUE;
END;
$$;
-- +goose StatementEnd

CREATE TABLE lucky_draws (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_by_member_id BIGINT NOT NULL REFERENCES group_members(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    title VARCHAR(200) NOT NULL,
    candidates JSONB NOT NULL,
    status lucky_draw_status NOT NULL DEFAULT 'DRAFT',
    result_index INTEGER,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_lucky_draws_title CHECK (BTRIM(title) <> ''),
    CONSTRAINT chk_lucky_draws_candidates CHECK (valid_lucky_draw_candidates(candidates)),
    CONSTRAINT chk_lucky_draws_result CHECK (
        (status = 'DRAFT' AND result_index IS NULL AND completed_at IS NULL)
        OR (status = 'COMPLETED' AND result_index IS NOT NULL AND completed_at IS NOT NULL
            AND result_index >= 0 AND result_index < jsonb_array_length(candidates))
    )
);

CREATE INDEX idx_lucky_draws_group_status ON lucky_draws (group_id, status, created_at DESC, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_lucky_draws_created_by_member_id ON lucky_draws (created_by_member_id);
CREATE INDEX idx_lucky_draws_deleted_at ON lucky_draws (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS lucky_draws;
DROP FUNCTION IF EXISTS valid_lucky_draw_candidates(JSONB);
DROP TYPE IF EXISTS lucky_draw_status;
