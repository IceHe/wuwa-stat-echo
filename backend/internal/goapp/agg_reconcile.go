package goapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (a *App) handleReconcileAggregates(w http.ResponseWriter, r *http.Request) {
	targetBits := parseInt64Default(r.URL.Query().Get("target_bits"), 0b11)
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)

	tuneCheck, err := a.reconcileTuneStatsAggregate(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to reconcile tune stats aggregate", 500))
		return
	}
	dcritCheck, err := a.reconcileEchoDcritAggregate(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to reconcile echo dcrit aggregate", 500))
		return
	}
	echoSummaryCheck, err := a.reconcileEchoSummaryAggregate(r.Context(), userID, targetBits)
	if err != nil {
		writeJSON(w, appError("failed to reconcile echo summary aggregate", 500))
		return
	}

	ok := tuneCheck["ok"] == true && dcritCheck["ok"] == true && echoSummaryCheck["ok"] == true
	writeJSON(w, success("reconcile aggregates", map[string]any{
		"ok":          ok,
		"user_id":     userID,
		"target_bits": targetBits,
		"checks": map[string]any{
			"tune_stats":   tuneCheck,
			"echo_dcrit":   dcritCheck,
			"echo_summary": echoSummaryCheck,
		},
	}))
}

func (a *App) reconcileTuneStatsAggregate(ctx context.Context) (map[string]any, error) {
	aggStats, err := a.loadTuneStatsFromAggregate(ctx, 0)
	if err != nil {
		return nil, err
	}
	rawStats, err := a.computeTuneStats(ctx, 0, 0, 0, 0, parseStatsWindow(""))
	if err != nil {
		return nil, err
	}
	aggJSON, err := json.Marshal(aggStats)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(rawStats)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":             bytes.Equal(aggJSON, rawJSON),
		"agg_data_total": aggStats.DataTotal,
		"raw_data_total": rawStats.DataTotal,
	}, nil
}

func (a *App) reconcileEchoDcritAggregate(ctx context.Context) (map[string]any, error) {
	aggData, err := a.loadEchoDcritFromAggregate(ctx, 0)
	if err != nil {
		return nil, err
	}
	rawData, err := a.computeEchoDcritRaw(ctx, 0, 0, 0, 0, parseStatsWindow(""))
	if err != nil {
		return nil, err
	}
	aggJSON, err := json.Marshal(aggData)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(rawData)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":              bytes.Equal(aggJSON, rawJSON),
		"agg_echo_count":  aggData["echo_count"],
		"raw_echo_count":  rawData["echo_count"],
		"agg_dcrit_total": aggData["dcrit_total"],
		"raw_dcrit_total": rawData["dcrit_total"],
	}, nil
}

func (a *App) reconcileEchoSummaryAggregate(ctx context.Context, userID int64, targetBits int64) (map[string]any, error) {
	aggData, err := a.loadEchoSummaryFromAggregate(ctx, userID, targetBits)
	if err != nil {
		return nil, err
	}
	rawData, err := a.computeEchoLogsAnalysisRaw(ctx, userID, targetBits)
	if err != nil {
		return nil, err
	}
	aggJSON, err := json.Marshal(aggData)
	if err != nil {
		return nil, err
	}
	rawJSON, err := json.Marshal(rawData)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":                          bytes.Equal(aggJSON, rawJSON),
		"user_id":                     userID,
		"target_bits":                 targetBits,
		"agg_target":                  aggData["target"],
		"raw_target":                  rawData["target"],
		"agg_target_echo_distance":    aggData["target_echo_distance"],
		"raw_target_echo_distance":    rawData["target_echo_distance"],
		"agg_target_substat_distance": aggData["target_substat_distance"],
		"raw_target_substat_distance": rawData["target_substat_distance"],
	}, nil
}

func (a *App) computeEchoDcritRaw(ctx context.Context, userID int64, size int, afterID int64, beforeID int64, window statsWindow) (map[string]any, error) {
	query := `select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0`
	args := []any{}
	arg := 1
	if userID > 0 {
		query += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, userID)
		arg++
	}
	if afterID > 0 {
		query += fmt.Sprintf(" and id > $%d", arg)
		args = append(args, afterID)
		arg++
	}
	if beforeID > 0 {
		query += fmt.Sprintf(" and id < $%d", arg)
		args = append(args, beforeID)
		arg++
	}
	if since := window.sinceTime(); since != nil {
		query += fmt.Sprintf(" and updated_at >= $%d", arg)
		args = append(args, *since)
		arg++
	}
	if effectiveSize := window.applyLimit(size); effectiveSize > 0 {
		query += fmt.Sprintf(" limit %d", effectiveSize)
	}
	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	echoCount := 0
	dcritTotal := 0
	counts := map[string]map[string]int{}
	for rows.Next() {
		var s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&s1, &s2, &s3, &s4, &s5); err != nil {
			return nil, err
		}
		echoCount++
		substatAll := s1 | s2 | s3 | s4 | s5
		if substatAll&0b11 != 0b11 {
			continue
		}
		dcritTotal++
		critRateNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b01)
		critDmgNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b10)
		rk := fmt.Sprint(bitPos(critRateNum))
		dk := fmt.Sprint(bitPos(critDmgNum))
		if _, ok := counts[rk]; !ok {
			counts[rk] = map[string]int{}
		}
		counts[rk][dk]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{
		"echo_count":       echoCount,
		"dcrit_total":      dcritTotal,
		"counts":           counts,
		"dcrit_rate_stats": newProportionStat(int64(dcritTotal), int64(echoCount)),
	}, nil
}

func (a *App) computeEchoLogsAnalysisRaw(ctx context.Context, userID int64, targetBits int64) (map[string]any, error) {
	query := `select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0`
	args := make([]any, 0, 1)
	if userID > 0 {
		query += ` and user_id = $1`
		args = append(args, userID)
	}
	query += ` order by updated_at desc`

	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EchoLog, 0)
	for rows.Next() {
		var s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&s1, &s2, &s3, &s4, &s5); err != nil {
			return nil, err
		}
		items = append(items, EchoLog{Substat1: s1, Substat2: s2, Substat3: s3, Substat4: s4, Substat5: s5})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return computeEchoLogsAnalysisFromItems(items, int64(len(items)), targetBits), nil
}
