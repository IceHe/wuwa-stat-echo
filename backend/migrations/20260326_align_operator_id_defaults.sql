BEGIN;

-- Backfill historical echo rows from their related tune logs before tightening constraints.
WITH inferred_operator AS (
    SELECT
        t.echo_id,
        MIN(t.operator_id) AS operator_id
    FROM public.wuwa_tune_log t
    WHERE t.operator_id IS NOT NULL
    GROUP BY t.echo_id
    HAVING COUNT(DISTINCT t.operator_id) = 1
)
UPDATE public.wuwa_echo_log e
SET operator_id = inferred_operator.operator_id
FROM inferred_operator
WHERE e.id = inferred_operator.echo_id
  AND e.operator_id IS NULL;

ALTER TABLE public.wuwa_echo_log
    ALTER COLUMN operator_id SET DEFAULT 1;

ALTER TABLE public.wuwa_tune_log
    ALTER COLUMN operator_id SET DEFAULT 1;

ALTER TABLE public.wuwa_echo_log
    ALTER COLUMN operator_id SET NOT NULL;

ALTER TABLE public.wuwa_tune_log
    ALTER COLUMN operator_id SET NOT NULL;

COMMIT;
