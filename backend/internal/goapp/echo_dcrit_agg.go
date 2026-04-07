package goapp

import (
	"context"
	"fmt"
	"net/http"
)

func (a *App) ensureEchoDcritAggregateReady(ctx context.Context) error {
	if _, err := a.db.Exec(ctx, `
		create table if not exists agg_echo_dcrit_counts (
			bucket_type text not null,
			bucket_key text not null,
			user_id bigint not null,
			crit_rate_tier integer not null,
			crit_dmg_tier integer not null,
			count bigint not null default 0,
			updated_at timestamptz not null default now(),
			primary key (bucket_type, bucket_key, user_id, crit_rate_tier, crit_dmg_tier)
		)
	`); err != nil {
		return err
	}

	var aggCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from agg_echo_dcrit_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll).Scan(&aggCount); err != nil {
		return err
	}
	if aggCount > 0 {
		return nil
	}

	var rawCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from wuwa_echo_log where deleted = 0`).Scan(&rawCount); err != nil {
		return err
	}
	if rawCount == 0 {
		return nil
	}

	return a.rebuildEchoDcritAggregate(ctx)
}

func (a *App) rebuildEchoDcritAggregate(ctx context.Context) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from agg_echo_dcrit_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll); err != nil {
		return fmt.Errorf("clear agg_echo_dcrit_counts: %w", err)
	}

	rows, err := tx.Query(ctx, `select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0`)
	if err != nil {
		return err
	}
	echoes, err := a.scanEchoLogs(rows)
	if err != nil {
		return err
	}
	for _, echo := range echoes {
		if err := a.applyEchoDcritDelta(ctx, tx, nil, &echo); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (a *App) handleRebuildEchoDcritAggregate(w http.ResponseWriter, r *http.Request) {
	job, err := a.runAggRebuildJob(r.Context(), "rebuild_echo_dcrit", func(ctx context.Context) error {
		return a.rebuildEchoDcritAggregate(ctx)
	})
	if err != nil {
		writeJSONWithStatus(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "failed to rebuild echo dcrit aggregate",
			"data":    map[string]any{"job": job},
		})
		return
	}
	writeJSON(w, success("rebuild echo dcrit aggregate", map[string]any{
		"bucket_type": aggBucketTypeAll,
		"bucket_key":  aggBucketKeyAll,
		"job":         job,
	}))
}

func (a *App) loadEchoDcritFromAggregate(ctx context.Context, userID int64) (map[string]any, error) {
	rows, err := a.db.Query(ctx, `
		select user_id, crit_rate_tier, crit_dmg_tier, count
		from agg_echo_dcrit_counts
		where bucket_type = $1 and bucket_key = $2 and user_id = $3
		order by crit_rate_tier, crit_dmg_tier
	`, aggBucketTypeAll, aggBucketKeyAll, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]map[string]int{}
	dcritTotal := 0
	for rows.Next() {
		var userID int64
		var critRateTier, critDmgTier int
		var count int64
		if err := rows.Scan(&userID, &critRateTier, &critDmgTier, &count); err != nil {
			return nil, err
		}
		rk := fmt.Sprint(critRateTier)
		dk := fmt.Sprint(critDmgTier)
		if _, ok := counts[rk]; !ok {
			counts[rk] = map[string]int{}
		}
		counts[rk][dk] += int(count)
		dcritTotal += int(count)
	}

	var echoCount int64
	countSQL := `select count(*) from wuwa_echo_log where deleted = 0`
	countArgs := []any{}
	if userID > 0 {
		countSQL += ` and user_id = $1`
		countArgs = append(countArgs, userID)
	}
	if err := a.db.QueryRow(ctx, countSQL, countArgs...).Scan(&echoCount); err != nil {
		return nil, err
	}

	return map[string]any{
		"echo_count":       echoCount,
		"dcrit_total":      dcritTotal,
		"counts":           counts,
		"dcrit_rate_stats": newProportionStat(int64(dcritTotal), echoCount),
	}, rows.Err()
}

func (a *App) applyEchoDcritDelta(ctx context.Context, q dbExecutor, before *EchoLog, after *EchoLog) error {
	if before != nil {
		if userID, rateTier, dmgTier, ok := echoDcritKey(*before); ok {
			if err := a.upsertEchoDcritCount(ctx, q, userID, rateTier, dmgTier, -1); err != nil {
				return err
			}
			if err := a.upsertEchoDcritCount(ctx, q, 0, rateTier, dmgTier, -1); err != nil {
				return err
			}
		}
	}
	if after != nil {
		if userID, rateTier, dmgTier, ok := echoDcritKey(*after); ok {
			if err := a.upsertEchoDcritCount(ctx, q, userID, rateTier, dmgTier, 1); err != nil {
				return err
			}
			if err := a.upsertEchoDcritCount(ctx, q, 0, rateTier, dmgTier, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) upsertEchoDcritCount(ctx context.Context, q dbExecutor, userID int64, critRateTier int, critDmgTier int, delta int64) error {
	_, err := q.Exec(ctx, `
		insert into agg_echo_dcrit_counts (bucket_type, bucket_key, user_id, crit_rate_tier, crit_dmg_tier, count, updated_at)
		values ($1, $2, $3, $4, $5, $6, now())
		on conflict (bucket_type, bucket_key, user_id, crit_rate_tier, crit_dmg_tier)
		do update set count = agg_echo_dcrit_counts.count + excluded.count, updated_at = now()
	`, aggBucketTypeAll, aggBucketKeyAll, userID, critRateTier, critDmgTier, delta)
	return err
}

func echoDcritKey(e EchoLog) (userID int64, critRateTier int, critDmgTier int, ok bool) {
	if e.Deleted != 0 {
		return 0, 0, 0, false
	}
	substats := []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5}
	substatAll := e.Substat1 | e.Substat2 | e.Substat3 | e.Substat4 | e.Substat5
	if substatAll&0b11 != 0b11 {
		return 0, 0, 0, false
	}
	critRateTier = bitPos(firstTierForMask(substats, 0b01))
	critDmgTier = bitPos(firstTierForMask(substats, 0b10))
	if critRateTier < 0 || critDmgTier < 0 {
		return 0, 0, 0, false
	}
	return e.UserID, critRateTier, critDmgTier, true
}
