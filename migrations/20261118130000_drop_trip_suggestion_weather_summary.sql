-- +goose Up
ALTER TABLE IF EXISTS trip_suggestions
    DROP COLUMN IF EXISTS weather_summary;

-- +goose Down
ALTER TABLE IF EXISTS trip_suggestions
    ADD COLUMN IF NOT EXISTS weather_summary TEXT;
