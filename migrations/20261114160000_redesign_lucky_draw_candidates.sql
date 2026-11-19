-- +goose Up
CREATE TYPE lucky_draw_type AS ENUM ('EQUAL', 'WEIGHTED');

ALTER TABLE lucky_draws
    DROP CONSTRAINT chk_lucky_draws_candidates,
    ADD COLUMN draw_type lucky_draw_type NOT NULL DEFAULT 'WEIGHTED';

-- +goose StatementBegin
CREATE FUNCTION migrate_lucky_draw_percentages(candidate_list JSONB) RETURNS JSONB
LANGUAGE plpgsql AS $$
DECLARE
    total NUMERIC;
    units INTEGER[] := ARRAY[]::INTEGER[];
    fractions NUMERIC[] := ARRAY[]::NUMERIC[];
    candidate JSONB;
    scaled NUMERIC;
    allocated INTEGER := 0;
    chosen INTEGER;
    position INTEGER := 0;
    result JSONB := '[]'::JSONB;
BEGIN
    SELECT SUM((item ->> 'weight')::NUMERIC)
    INTO total FROM jsonb_array_elements(candidate_list) item;
    FOR candidate IN SELECT * FROM jsonb_array_elements(candidate_list) LOOP
        position := position + 1;
        scaled := (candidate ->> 'weight')::NUMERIC * 10000 / total;
        units := array_append(units, GREATEST(1, FLOOR(scaled)::INTEGER));
        fractions := array_append(fractions, scaled - FLOOR(scaled));
        allocated := allocated + units[position];
    END LOOP;
    WHILE allocated < 10000 LOOP
        SELECT i INTO chosen FROM generate_subscripts(units, 1) i
        ORDER BY fractions[i] DESC, i LIMIT 1;
        units[chosen] := units[chosen] + 1;
        fractions[chosen] := -1;
        allocated := allocated + 1;
    END LOOP;
    WHILE allocated > 10000 LOOP
        SELECT i INTO chosen FROM generate_subscripts(units, 1) i
        WHERE units[i] > 1 ORDER BY units[i] DESC, i LIMIT 1;
        units[chosen] := units[chosen] - 1;
        allocated := allocated - 1;
    END LOOP;
    FOR position IN 1..array_length(units, 1) LOOP
        result := result || jsonb_build_array(jsonb_build_object(
            'type', 'TEXT', 'value', candidate_list -> (position - 1) ->> 'title',
            'weight', ROUND(units[position]::NUMERIC / 100, 2)
        ));
    END LOOP;
    RETURN result;
END;
$$;
-- +goose StatementEnd

UPDATE lucky_draws
SET draw_type = CASE WHEN (
        SELECT COUNT(DISTINCT item ->> 'weight') = 1
        FROM jsonb_array_elements(candidates) item
    ) THEN 'EQUAL'::lucky_draw_type ELSE 'WEIGHTED'::lucky_draw_type END,
    candidates = migrate_lucky_draw_percentages(candidates);

DROP FUNCTION migrate_lucky_draw_percentages(JSONB);

WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY created_at DESC, id DESC) AS position
    FROM lucky_draws WHERE status = 'DRAFT' AND deleted_at IS NULL
)
UPDATE lucky_draws SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
FROM ranked WHERE lucky_draws.id = ranked.id AND ranked.position > 1;

-- +goose StatementBegin
CREATE FUNCTION valid_lucky_draw_configuration(candidate_list JSONB, mode lucky_draw_type) RETURNS BOOLEAN
LANGUAGE plpgsql IMMUTABLE AS $$
DECLARE
    candidate JSONB;
    candidate_kind TEXT;
    candidate_key TEXT;
    percentage NUMERIC;
    total NUMERIC := 0;
    position INTEGER := 0;
    count INTEGER;
    expected INTEGER;
    seen TEXT[] := ARRAY[]::TEXT[];
BEGIN
    IF candidate_list IS NULL OR jsonb_typeof(candidate_list) <> 'array' OR mode IS NULL THEN
        RETURN FALSE;
    END IF;
    count := jsonb_array_length(candidate_list);
    IF count < 2 OR count > 100 THEN
        RETURN FALSE;
    END IF;
    FOR candidate IN SELECT * FROM jsonb_array_elements(candidate_list) LOOP
        IF jsonb_typeof(candidate) <> 'object' THEN RETURN FALSE; END IF;
        candidate_kind := candidate ->> 'type';
        IF candidate_kind = 'TEXT' THEN
            IF jsonb_typeof(candidate -> 'value') IS DISTINCT FROM 'string'
                OR BTRIM(candidate ->> 'value') = ''
                OR CHAR_LENGTH(candidate ->> 'value') > 255 THEN RETURN FALSE; END IF;
            candidate_key := 'TEXT:' || LOWER(BTRIM(candidate ->> 'value'));
        ELSIF candidate_kind = 'USER' THEN
            IF jsonb_typeof(candidate -> 'value') IS DISTINCT FROM 'number'
                OR (candidate ->> 'value') !~ '^[1-9][0-9]*$'
                OR (candidate ->> 'value')::NUMERIC > 9223372036854775807 THEN RETURN FALSE; END IF;
            candidate_key := 'USER:' || (candidate ->> 'value');
        ELSE
            RETURN FALSE;
        END IF;
        IF candidate_key = ANY(seen) THEN RETURN FALSE; END IF;
        seen := array_append(seen, candidate_key);
        IF jsonb_typeof(candidate -> 'weight') IS DISTINCT FROM 'number' THEN RETURN FALSE; END IF;
        percentage := (candidate ->> 'weight')::NUMERIC;
        IF percentage <= 0 OR percentage > 100 OR SCALE(percentage) > 2 OR percentage * 100 <> TRUNC(percentage * 100) THEN RETURN FALSE; END IF;
        IF mode = 'EQUAL' THEN
            expected := 10000 / count + CASE WHEN position < 10000 % count THEN 1 ELSE 0 END;
            IF percentage * 100 <> expected THEN RETURN FALSE; END IF;
        END IF;
        position := position + 1;
        total := total + percentage;
    END LOOP;
    RETURN total = 100;
END;
$$;
-- +goose StatementEnd

ALTER TABLE lucky_draws
    ALTER COLUMN draw_type DROP DEFAULT,
    ADD CONSTRAINT chk_lucky_draws_configuration CHECK (valid_lucky_draw_configuration(candidates, draw_type));

CREATE UNIQUE INDEX idx_lucky_draws_one_draft_per_group ON lucky_draws (group_id)
WHERE status = 'DRAFT' AND deleted_at IS NULL;

CREATE INDEX idx_lucky_draws_history ON lucky_draws (group_id, completed_at DESC, id DESC)
WHERE status = 'COMPLETED' AND deleted_at IS NULL;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM lucky_draws, jsonb_array_elements(candidates) candidate
        WHERE candidate ->> 'type' = 'USER'
    ) THEN
        RAISE EXCEPTION 'cannot roll back typed user candidates to text without an explicit data conversion';
    END IF;
END;
$$;
-- +goose StatementEnd

DROP INDEX idx_lucky_draws_history;
DROP INDEX idx_lucky_draws_one_draft_per_group;
ALTER TABLE lucky_draws DROP CONSTRAINT chk_lucky_draws_configuration;

UPDATE lucky_draws SET candidates = (
    SELECT jsonb_agg(jsonb_build_object('title', candidate ->> 'value',
        'weight', ((candidate ->> 'weight')::NUMERIC * 100)::BIGINT) ORDER BY position)
    FROM jsonb_array_elements(candidates) WITH ORDINALITY AS options(candidate, position)
);

ALTER TABLE lucky_draws
    DROP COLUMN draw_type,
    ADD CONSTRAINT chk_lucky_draws_candidates CHECK (valid_lucky_draw_candidates(candidates));

DROP FUNCTION valid_lucky_draw_configuration(JSONB, lucky_draw_type);
DROP TYPE lucky_draw_type;
