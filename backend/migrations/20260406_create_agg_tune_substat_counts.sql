BEGIN;

CREATE TABLE IF NOT EXISTS public.agg_tune_substat_counts (
    bucket_type TEXT NOT NULL,
    bucket_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    substat INTEGER NOT NULL,
    value INTEGER NOT NULL,
    position INTEGER NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_type, bucket_key, user_id, substat, value, position)
);

CREATE INDEX IF NOT EXISTS idx_agg_tune_substat_counts_lookup
    ON public.agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, position);

COMMIT;
