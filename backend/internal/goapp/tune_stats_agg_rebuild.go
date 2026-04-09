package goapp

import (
	"context"
	"fmt"
	"net/http"
)

func (a *App) ensureTuneStatsAggregateReady(ctx context.Context) error {
	if _, err := a.db.Exec(ctx, `
		create table if not exists agg_tune_substat_counts (
			bucket_type text not null,
			bucket_key text not null,
			user_id bigint not null,
			substat integer not null,
			value integer not null,
			position integer not null,
			count bigint not null default 0,
			updated_at timestamptz not null default now(),
			primary key (bucket_type, bucket_key, user_id, substat, value, position)
		)
	`); err != nil {
		return err
	}

	var aggCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from agg_tune_substat_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll).Scan(&aggCount); err != nil {
		return err
	}
	if aggCount > 0 {
		return nil
	}

	var rawCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from wuwa_tune_log where deleted = 0`).Scan(&rawCount); err != nil {
		return err
	}
	if rawCount == 0 {
		return nil
	}

	return a.rebuildTuneStatsAggregate(ctx)
}

func (a *App) rebuildTuneStatsAggregate(ctx context.Context) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from agg_tune_substat_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll); err != nil {
		return fmt.Errorf("clear agg_tune_substat_counts: %w", err)
	}

	insertStatements := []struct {
		sql  string
		args []any
	}{
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, user_id, substat, value, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by user_id, substat, value, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, user_id, substat, $3, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by user_id, substat, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll, aggValueAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, 0, substat, value, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by substat, value, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, 0, substat, $3, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by substat, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll, aggValueAll},
		},
	}

	for index, statement := range insertStatements {
		if _, err := tx.Exec(ctx, statement.sql, statement.args...); err != nil {
			return fmt.Errorf("rebuild agg_tune_substat_counts step %d: %w", index+1, err)
		}
	}

	return tx.Commit(ctx)
}

func (a *App) handleRebuildTuneStatsAggregate(w http.ResponseWriter, r *http.Request) {
	job, err := a.runAggRebuildJob(r.Context(), "rebuild_tune_stats", func(ctx context.Context) error {
		if err := a.rebuildTuneStatsAggregate(ctx); err != nil {
			return err
		}
		return a.refreshCachedTuneStats(ctx)
	})
	if err != nil {
		writeJSONWithStatus(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "failed to rebuild tune stats aggregate",
			"data":    map[string]any{"job": job},
		})
		return
	}
	writeJSON(w, success("rebuild tune stats aggregate", map[string]any{
		"bucket_type": aggBucketTypeAll,
		"bucket_key":  aggBucketKeyAll,
		"job":         job,
	}))
}

func (a *App) applyTuneStatsDelta(ctx context.Context, q dbExecutor, logs []SubstatLog, delta int64) error {
	if len(logs) == 0 || delta == 0 {
		return nil
	}

	statement := `
		insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, now())
		on conflict (bucket_type, bucket_key, user_id, substat, value, position)
		do update set count = agg_tune_substat_counts.count + excluded.count, updated_at = now()
	`

	for _, logItem := range logs {
		if logItem.Substat < 0 || logItem.Substat >= len(substatDefs) {
			continue
		}
		if logItem.Position < 0 || logItem.Position >= 5 {
			continue
		}

		rows := [][4]int64{
			{logItem.UserID, int64(logItem.Substat), int64(logItem.Value), int64(logItem.Position)},
			{logItem.UserID, int64(logItem.Substat), aggValueAll, int64(logItem.Position)},
			{0, int64(logItem.Substat), int64(logItem.Value), int64(logItem.Position)},
			{0, int64(logItem.Substat), aggValueAll, int64(logItem.Position)},
		}
		for _, row := range rows {
			if _, err := q.Exec(ctx, statement, aggBucketTypeAll, aggBucketKeyAll, row[0], row[1], row[2], row[3], delta); err != nil {
				return err
			}
		}
	}

	return nil
}
