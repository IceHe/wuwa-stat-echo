BEGIN;

CREATE TABLE IF NOT EXISTS public.agg_echo_summary (
    bucket_type TEXT NOT NULL,
    bucket_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    target_bits BIGINT NOT NULL,
    echo_count BIGINT NOT NULL DEFAULT 0,
    hit_count BIGINT NOT NULL DEFAULT 0,
    substat_total BIGINT NOT NULL DEFAULT 0,
    exp_total BIGINT NOT NULL DEFAULT 0,
    exp_recycled BIGINT NOT NULL DEFAULT 0,
    tuner_recycled BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_type, bucket_key, user_id, target_bits)
);

CREATE INDEX IF NOT EXISTS idx_agg_echo_summary_lookup
    ON public.agg_echo_summary (bucket_type, bucket_key, user_id, target_bits);

COMMIT;
