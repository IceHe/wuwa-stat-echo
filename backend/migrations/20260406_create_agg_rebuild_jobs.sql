BEGIN;

CREATE TABLE IF NOT EXISTS public.agg_rebuild_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    message TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_agg_rebuild_jobs_job_type_status
    ON public.agg_rebuild_jobs (job_type, status, started_at DESC);

COMMIT;
