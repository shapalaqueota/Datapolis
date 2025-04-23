-- +goose Up

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE geo_collections (
                                 id          SERIAL PRIMARY KEY,
                                 name        TEXT        NOT NULL,
                                 description TEXT,
                                 srid        INT         NOT NULL DEFAULT 4326,
                                 user_id     INT         NOT NULL,
                                 created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                 updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE geo_features (
                              id            SERIAL PRIMARY KEY,
                              collection_id INT REFERENCES geo_collections(id) ON DELETE CASCADE,
                              properties    JSONB       DEFAULT '{}'::jsonb,
                              geometry      geometry    NOT NULL,
                              created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX geo_features_geom_gix ON geo_features USING GIST(geometry);

-- +goose Down

DROP INDEX IF EXISTS geo_features_geom_gix;
DROP TABLE IF EXISTS geo_features;
DROP TABLE IF EXISTS geo_collections;
DROP EXTENSION IF EXISTS postgis;
