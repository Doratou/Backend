-- +goose Up
CREATE TABLE IF NOT EXISTS token_prices (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    price_toman BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO token_prices (id, price_toman) VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS token_prices;
