BEGIN;

CREATE TABLE IF NOT EXISTS public.agg_echo_dcrit_counts (
    bucket_type TEXT NOT NULL,
    bucket_key TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    crit_rate_tier INTEGER NOT NULL,
    crit_dmg_tier INTEGER NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_type, bucket_key, user_id, crit_rate_tier, crit_dmg_tier)
);

CREATE INDEX IF NOT EXISTS idx_agg_echo_dcrit_counts_lookup
    ON public.agg_echo_dcrit_counts (bucket_type, bucket_key, user_id);

COMMIT;
