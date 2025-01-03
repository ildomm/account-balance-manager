DROP TYPE IF EXISTS game_statuses;
CREATE TYPE game_statuses AS ENUM ('win', 'lost');

DROP TYPE IF EXISTS transaction_sources;
CREATE TYPE transaction_sources AS ENUM ('game', 'server', 'payment');

CREATE TABLE IF NOT EXISTS game_results (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              BIGINT NOT NULL, /* DO NOT make it a Referential Integrity Constraint for performance reasons ONLY */
    game_status          game_statuses NOT NULL,
    transaction_source   transaction_sources NOT NULL,
    transaction_id       VARCHAR NOT NULL,
    amount               DECIMAL(10,2) NOT NULL,
    created_at           DATE NOT NULL
);