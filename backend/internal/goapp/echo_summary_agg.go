package goapp

import (
	"context"
	"fmt"
	"net/http"
)

type echoSummaryAggKey struct {
	userID     int64
	targetBits int64
}

type echoSummaryAggValue struct {
	echoCount     int64
	hitCount      int64
	substatTotal  int64
	expTotal      int64
	expRecycled   int64
	tunerRecycled int64
}

func (a *App) ensureEchoSummaryAggregateReady(ctx context.Context) error {
	if _, err := a.db.Exec(ctx, `
		create table if not exists agg_echo_summary (
			bucket_type text not null,
			bucket_key text not null,
			user_id bigint not null,
			target_bits bigint not null,
			echo_count bigint not null default 0,
			hit_count bigint not null default 0,
			substat_total bigint not null default 0,
			exp_total bigint not null default 0,
			exp_recycled bigint not null default 0,
			tuner_recycled bigint not null default 0,
			updated_at timestamptz not null default now(),
			primary key (bucket_type, bucket_key, user_id, target_bits)
		)
	`); err != nil {
		return err
	}

	var aggCount int64
	if err := a.db.QueryRow(ctx, `
		select count(*)
		from agg_echo_summary
		where bucket_type = $1 and bucket_key = $2 and user_id = 0 and target_bits = 0
	`, aggBucketTypeAll, aggBucketKeyAll).Scan(&aggCount); err != nil {
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

	return a.rebuildEchoSummaryAggregate(ctx)
}

func (a *App) rebuildEchoSummaryAggregate(ctx context.Context) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from agg_echo_summary where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll); err != nil {
		return fmt.Errorf("clear agg_echo_summary: %w", err)
	}

	rows, err := tx.Query(ctx, `select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0`)
	if err != nil {
		return err
	}
	echoes, err := a.scanEchoLogs(rows)
	if err != nil {
		return err
	}

	aggRows := map[echoSummaryAggKey]echoSummaryAggValue{}
	for _, echo := range echoes {
		for key, value := range echoSummaryRowsForEcho(echo, 1) {
			current := aggRows[key]
			current.echoCount += value.echoCount
			current.hitCount += value.hitCount
			current.substatTotal += value.substatTotal
			current.expTotal += value.expTotal
			current.expRecycled += value.expRecycled
			current.tunerRecycled += value.tunerRecycled
			aggRows[key] = current
		}
	}

	for key, value := range aggRows {
		if err := a.upsertEchoSummaryCount(ctx, tx, key.userID, key.targetBits, value); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (a *App) handleRebuildEchoSummaryAggregate(w http.ResponseWriter, r *http.Request) {
	job, err := a.runAggRebuildJob(r.Context(), "rebuild_echo_summary", func(ctx context.Context) error {
		return a.rebuildEchoSummaryAggregate(ctx)
	})
	if err != nil {
		writeJSONWithStatus(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "failed to rebuild echo summary aggregate",
			"data":    map[string]any{"job": job},
		})
		return
	}
	writeJSON(w, success("rebuild echo summary aggregate", map[string]any{
		"bucket_type": aggBucketTypeAll,
		"bucket_key":  aggBucketKeyAll,
		"job":         job,
	}))
}

func (a *App) loadEchoSummaryFromAggregate(ctx context.Context, userID int64, targetBits int64) (map[string]any, error) {
	query := `
		select target_bits, echo_count, hit_count, substat_total, exp_total, exp_recycled, tuner_recycled
		from agg_echo_summary
		where bucket_type = $1 and bucket_key = $2 and user_id = $3 and target_bits = any($4)
	`
	targetList := []int64{0}
	if targetBits != 0 {
		targetList = append(targetList, targetBits)
	}
	rows, err := a.db.Query(ctx, query, aggBucketTypeAll, aggBucketKeyAll, userID, targetList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalRow := echoSummaryAggValue{}
	targetRow := echoSummaryAggValue{}
	foundTotal := false
	for rows.Next() {
		var rowTargetBits int64
		var item echoSummaryAggValue
		if err := rows.Scan(&rowTargetBits, &item.echoCount, &item.hitCount, &item.substatTotal, &item.expTotal, &item.expRecycled, &item.tunerRecycled); err != nil {
			return nil, err
		}
		if rowTargetBits == 0 {
			totalRow = item
			foundTotal = true
			continue
		}
		if rowTargetBits == targetBits {
			targetRow = item
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !foundTotal {
		return nil, nil
	}

	targetCount := targetRow.hitCount
	matchedExpRecycled := targetRow.expRecycled
	matchedTunerRecycled := targetRow.tunerRecycled
	if targetBits == 0 {
		targetCount = totalRow.echoCount
		matchedExpRecycled = totalRow.expRecycled
		matchedTunerRecycled = totalRow.tunerRecycled
	}

	targetEchoDistance, targetSubstatDistance, err := a.computeEchoAnalysisDistance(ctx, userID, targetBits)
	if err != nil {
		return nil, err
	}

	expConsumed := int(mathCeilDiv(totalRow.expTotal-(totalRow.expRecycled-matchedExpRecycled), expGold))
	tunerConsumed := int(totalRow.substatTotal*10 - (totalRow.tunerRecycled - matchedTunerRecycled))
	resp := map[string]any{
		"sample_size":             totalRow.echoCount,
		"target_echo_distance":    targetEchoDistance,
		"target_substat_distance": targetSubstatDistance,
		"target":                  targetCount,
		"target_avg_echo":         0.0,
		"target_avg_substat":      0.0,
		"tuner_consumed":          tunerConsumed,
		"tuner_consumed_avg":      0.0,
		"exp_consumed":            expConsumed,
		"exp_consumed_avg":        0.0,
		"target_rate_stats":       newProportionStat(targetCount, totalRow.echoCount),
	}
	if targetCount > 0 {
		resp["target_avg_echo"] = rounded(float64(totalRow.echoCount)/float64(targetCount), 1)
		resp["target_avg_substat"] = rounded(float64(totalRow.substatTotal)/float64(targetCount), 1)
		resp["tuner_consumed_avg"] = int(mathCeilDiv(int64(tunerConsumed), targetCount))
		resp["exp_consumed_avg"] = int(mathCeilDiv(int64(expConsumed), targetCount))
	}
	return resp, nil
}

func (a *App) computeEchoAnalysisDistance(ctx context.Context, userID int64, targetBits int64) (int, int, error) {
	query := `select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0`
	args := []any{}
	if userID > 0 {
		query += ` and user_id = $1`
		args = append(args, userID)
	}
	query += ` order by updated_at desc`

	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	found := false
	targetEchoDistance := 0
	targetSubstatDistance := 0
	substatTotal := 0
	index := 0
	for rows.Next() {
		var s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&s1, &s2, &s3, &s4, &s5); err != nil {
			return 0, 0, err
		}
		substatAll := (s1 | s2 | s3 | s4 | s5) & substatMask
		substatTotal += bitCount(substatAll)
		if substatAll&targetBits == targetBits && !found {
			found = true
			targetEchoDistance = index
			targetSubstatDistance = substatTotal
		}
		index++
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	if !found {
		targetEchoDistance = index
		targetSubstatDistance = substatTotal
	}
	return targetEchoDistance, targetSubstatDistance, nil
}

func (a *App) applyEchoSummaryDelta(ctx context.Context, q dbExecutor, before *EchoLog, after *EchoLog) error {
	if before != nil {
		for key, value := range echoSummaryRowsForEcho(*before, -1) {
			if err := a.upsertEchoSummaryCount(ctx, q, key.userID, key.targetBits, value); err != nil {
				return err
			}
		}
	}
	if after != nil {
		for key, value := range echoSummaryRowsForEcho(*after, 1) {
			if err := a.upsertEchoSummaryCount(ctx, q, key.userID, key.targetBits, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) upsertEchoSummaryCount(ctx context.Context, q dbExecutor, userID int64, targetBits int64, delta echoSummaryAggValue) error {
	if delta.echoCount == 0 && delta.hitCount == 0 && delta.substatTotal == 0 && delta.expTotal == 0 && delta.expRecycled == 0 && delta.tunerRecycled == 0 {
		return nil
	}
	_, err := q.Exec(ctx, `
		insert into agg_echo_summary (
			bucket_type, bucket_key, user_id, target_bits,
			echo_count, hit_count, substat_total, exp_total, exp_recycled, tuner_recycled, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		on conflict (bucket_type, bucket_key, user_id, target_bits)
		do update set
			echo_count = agg_echo_summary.echo_count + excluded.echo_count,
			hit_count = agg_echo_summary.hit_count + excluded.hit_count,
			substat_total = agg_echo_summary.substat_total + excluded.substat_total,
			exp_total = agg_echo_summary.exp_total + excluded.exp_total,
			exp_recycled = agg_echo_summary.exp_recycled + excluded.exp_recycled,
			tuner_recycled = agg_echo_summary.tuner_recycled + excluded.tuner_recycled,
			updated_at = now()
	`, aggBucketTypeAll, aggBucketKeyAll, userID, targetBits, delta.echoCount, delta.hitCount, delta.substatTotal, delta.expTotal, delta.expRecycled, delta.tunerRecycled)
	return err
}

func echoSummaryRowsForEcho(e EchoLog, delta int64) map[echoSummaryAggKey]echoSummaryAggValue {
	if e.Deleted != 0 || delta == 0 {
		return nil
	}

	substatBits := (e.Substat1 | e.Substat2 | e.Substat3 | e.Substat4 | e.Substat5) & substatMask
	substatCount := int64(bitCount(substatBits))
	expTotal := int64(expTable[0][int(substatCount)])
	expRecycled := int64(expReturn[int(substatCount)])
	tunerRecycled := substatCount * tunerRecycledPerSubstat

	rows := map[echoSummaryAggKey]echoSummaryAggValue{}
	for _, userID := range echoSummaryUserBuckets(e.UserID) {
		rows[echoSummaryAggKey{userID: userID, targetBits: 0}] = echoSummaryAggValue{
			echoCount:     delta,
			substatTotal:  delta * substatCount,
			expTotal:      delta * expTotal,
			expRecycled:   delta * expRecycled,
			tunerRecycled: delta * tunerRecycled,
		}
		for _, subsetBits := range expandTargetBits(substatBits) {
			rows[echoSummaryAggKey{userID: userID, targetBits: subsetBits}] = echoSummaryAggValue{
				hitCount:      delta,
				substatTotal:  delta * substatCount,
				expTotal:      delta * expTotal,
				expRecycled:   delta * expRecycled,
				tunerRecycled: delta * tunerRecycled,
			}
		}
	}
	return rows
}

func echoSummaryUserBuckets(userID int64) []int64 {
	if userID == 0 {
		return []int64{0}
	}
	return []int64{userID, 0}
}

func expandTargetBits(bits int64) []int64 {
	if bits == 0 {
		return nil
	}
	positions := []int{}
	for pos := 0; pos < substatBitWidth; pos++ {
		if bits&(1<<pos) != 0 {
			positions = append(positions, pos)
		}
	}
	out := make([]int64, 0, (1<<len(positions))-1)
	for mask := 1; mask < (1 << len(positions)); mask++ {
		var subset int64
		for i, pos := range positions {
			if mask&(1<<i) != 0 {
				subset |= 1 << pos
			}
		}
		out = append(out, subset)
	}
	return out
}

func mathCeilDiv(dividend int64, divisor int64) int64 {
	if divisor <= 0 {
		return 0
	}
	if dividend <= 0 {
		return 0
	}
	return (dividend + divisor - 1) / divisor
}
