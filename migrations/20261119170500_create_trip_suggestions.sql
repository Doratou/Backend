-- +goose Up
CREATE TABLE trip_suggestions (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    latitude DOUBLE PRECISION NOT NULL CHECK (latitude >= -90 AND latitude <= 90),
    longitude DOUBLE PRECISION NOT NULL CHECK (longitude >= -180 AND longitude <= 180),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    people_count SMALLINT NOT NULL CHECK (people_count >= 1 AND people_count <= 100),
    input JSONB NOT NULL,
    weather_snapshot JSONB,
    summary TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL,
    prompt_version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_trip_suggestions_date_range CHECK (end_date >= start_date)
);

CREATE TABLE trip_suggestion_items (
    id BIGSERIAL PRIMARY KEY,
    suggestion_id BIGINT NOT NULL REFERENCES trip_suggestions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (BTRIM(name) <> ''),
    priority TEXT NOT NULL CHECK (priority IN ('ضروری', 'پیشنهادی', 'اختیاری')),
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_trip_suggestions_group_created_at
    ON trip_suggestions (group_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_trip_suggestions_created_by_user_id
    ON trip_suggestions (created_by_user_id);
CREATE INDEX idx_trip_suggestion_items_suggestion_id
    ON trip_suggestion_items (suggestion_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS trip_suggestion_items;
DROP TABLE IF EXISTS trip_suggestions;
