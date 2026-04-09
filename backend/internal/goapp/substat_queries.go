package goapp

import (
	"fmt"
	"net/http"
	"strconv"
)

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"message": "Hello, Wuwa!"})
}

func (a *App) handleListSubstatLogs(w http.ResponseWriter, r *http.Request) {
	pageNum := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	offset := (pageNum - 1) * pageSize
	rows, err := a.db.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log order by id desc offset $1 limit $2", offset, pageSize)
	if err != nil {
		writeJSON(w, appError("failed to get tune logs", 500))
		return
	}
	defer rows.Close()
	var data []SubstatLog
	for rows.Next() {
		var item SubstatLog
		if err := rows.Scan(&item.ID, &item.Substat, &item.Value, &item.Position, &item.EchoID, &item.UserID, &item.OperatorID, &item.Timestamp, &item.Deleted); err != nil {
			writeJSON(w, appError("failed to get tune logs", 500))
			return
		}
		data = append(data, item)
	}
	var total int64
	if err := a.db.QueryRow(r.Context(), "select count(id) from wuwa_tune_log").Scan(&total); err != nil {
		writeJSON(w, appError("failed to get tune logs", 500))
		return
	}
	writeJSON(w, page("tune logs", data, total, pageNum, pageSize))
}

func (a *App) handleTuneStats(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	afterID := parseInt64Default(r.URL.Query().Get("after_id"), 0)
	beforeID := parseInt64Default(r.URL.Query().Get("before_id"), 0)
	window := parseStatsWindow(r.URL.Query().Get("window"))

	var (
		stats *TuneStatsResponse
		err   error
	)
	if window.isAll() && size == 0 && afterID == 0 && beforeID == 0 {
		stats, err = a.loadTuneStatsFromAggregate(r.Context(), userID)
	}
	if stats == nil && err == nil {
		stats, err = a.computeTuneStats(r.Context(), size, userID, afterID, beforeID, window)
	}
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	if stats != nil {
		stats.Window = window.Name
	}
	if userID > 0 {
		var globalStats *TuneStatsResponse
		if window.isAll() && size == 0 && afterID == 0 && beforeID == 0 {
			globalStats, err = a.loadTuneStatsFromAggregate(r.Context(), 0)
		} else {
			globalStats, err = a.computeTuneStats(r.Context(), size, 0, afterID, beforeID, window)
		}
		if err != nil {
			writeJSON(w, appError("failed to get stats", 500))
			return
		}
		stats.BaselineCompare = buildTuneStatsBaselineCompare(stats, globalStats)
	}
	writeJSON(w, success("tune stats", stats))
}

func (a *App) handleSubstatDistanceAnalysis(w http.ResponseWriter, r *http.Request) {
	stats, err := a.computeTuneStats(r.Context(), parseIntDefault(r.URL.Query().Get("size"), 0), 0, 0, 0, parseStatsWindow(""))
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	writeJSON(w, success("tune stats", map[string]any{"data_total": stats.DataTotal, "substat_dict": stats.SubstatDict, "substat_distance": stats.SubstatDistance}))
}

func (a *App) handleAnalyzeEcho(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to get echo log", 500))
		return
	}
	stats := a.fillCurrentPositionPercent(payload, a.getCachedTuneStats())
	if stats == nil {
		stats = &TuneStatsResponse{SubstatDict: newSubstatDict(), PositionTotal: make([]int, 5), SubstatPosTotal: make([][]int, 13)}
	}
	resonator := r.URL.Query().Get("resonator")
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	stats.ResonatorTemplate = &template
	stats.Score = scoreEcho(payload, resonator, r.URL.Query().Get("cost"))
	pos := currentPos(payload)
	critCount := bitCount(payload.SubstatAll & 0b11)
	if pos >= 0 && pos < len(twoCritPercent) && critCount >= 0 && critCount < len(twoCritPercent[pos]) {
		stats.TwoCritPercent = twoCritPercent[pos][critCount]
	}
	writeJSON(w, success("echo log", stats))
}

func (a *App) handleEchoDcrit(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	beforeID := parseInt64Default(r.URL.Query().Get("before_id"), 0)
	afterID := parseInt64Default(r.URL.Query().Get("after_id"), 0)
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	window := parseStatsWindow(r.URL.Query().Get("window"))
	if window.isAll() && size == 0 && beforeID == 0 && afterID == 0 {
		if data, err := a.loadEchoDcritFromAggregate(r.Context(), userID); err == nil && data != nil {
			if userID > 0 {
				if globalData, globalErr := a.loadEchoDcritFromAggregate(r.Context(), 0); globalErr == nil && globalData != nil {
					data["baseline_compare"] = map[string]any{
						"dcrit_rate": buildRateComparison(
							data["dcrit_rate_stats"].(*ProportionStat),
							globalData["dcrit_rate_stats"].(*ProportionStat),
						),
					}
				}
			}
			data["window"] = window.Name
			writeJSON(w, success("test", data))
			return
		}
	}
	query := "select id, substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	var args []any
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
	rows, err := a.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to test", 500))
		return
	}
	defer rows.Close()
	echoCount := 0
	dcritTotal := 0
	counts := map[string]map[string]int{}
	for rows.Next() {
		echoCount++
		var id, s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&id, &s1, &s2, &s3, &s4, &s5); err != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		substatAll := s1 | s2 | s3 | s4 | s5
		if substatAll&0b11 == 0b11 {
			dcritTotal++
			critRateNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b01)
			critDmgNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b10)
			rk := strconv.Itoa(bitPos(critRateNum))
			dk := strconv.Itoa(bitPos(critDmgNum))
			if _, ok := counts[rk]; !ok {
				counts[rk] = map[string]int{}
			}
			counts[rk][dk]++
		}
	}
	resp := map[string]any{
		"echo_count":       echoCount,
		"dcrit_total":      dcritTotal,
		"counts":           counts,
		"dcrit_rate_stats": newProportionStat(int64(dcritTotal), int64(echoCount)),
		"window":           window.Name,
	}
	if userID > 0 {
		globalResp, globalErr := a.computeEchoDcritRaw(r.Context(), 0, size, afterID, beforeID, window)
		if globalErr != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		resp["baseline_compare"] = map[string]any{
			"dcrit_rate": buildRateComparison(resp["dcrit_rate_stats"].(*ProportionStat), globalResp["dcrit_rate_stats"].(*ProportionStat)),
		}
	}
	writeJSON(w, success("test", resp))
}

func (a *App) handleTestZero(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	query := "select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	if size > 0 {
		query += fmt.Sprintf(" limit %d", size)
	}
	rows, err := a.db.Query(r.Context(), query)
	if err != nil {
		writeJSON(w, appError("failed to test", 500))
		return
	}
	defer rows.Close()
	echoCount, dcritTotal, dcrit2, dcrit3, dcrit4 := 0, 0, 0, 0, 0
	for rows.Next() {
		var s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&s1, &s2, &s3, &s4, &s5); err != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		echoCount++
		substatAll := s1 | s2 | s3 | s4 | s5
		if substatAll&0b11 == 0b11 {
			dcritTotal++
			if (s1|s2)&0b11 == 0b11 {
				dcrit2++
			}
			if (s1|s2|s3)&0b11 == 0b11 {
				dcrit3++
			}
			if (s1|s2|s3|s4)&0b11 == 0b11 {
				dcrit4++
			}
		}
	}
	rate := func(v int) string {
		if echoCount == 0 {
			return "0%"
		}
		return fmt.Sprintf("%v%%", float64(v)/float64(echoCount)*100)
	}
	perEcho := func(v int) string {
		if echoCount == 0 || v == 0 {
			return "0"
		}
		return fmt.Sprintf("%v", 1.0/(float64(v)/float64(echoCount)))
	}
	writeJSON(w, success("test", map[string]any{
		"echo_count":        echoCount,
		"dcrit_total":       dcritTotal,
		"dcrit2_total":      dcrit2,
		"dcrit3_total":      dcrit3,
		"dcrit4_total":      dcrit4,
		"dcrit2_rate":       rate(dcrit2),
		"dcrit3_rate":       rate(dcrit3),
		"dcrit4_rate":       rate(dcrit4),
		"dcrit2_per_echoes": perEcho(dcrit2),
		"dcrit3_per_echoes": perEcho(dcrit3),
		"dcrit4_per_echoes": perEcho(dcrit4),
	}))
}

func (a *App) handlePredictEchoSubstat(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	target := payload.SubstatAll & substatMask
	if target < 0 {
		writeJSON(w, appError("substat_target must >= 0", 500))
		return
	}
	if bitPos(target) >= 5 {
		writeJSON(w, success("predict echo substat", map[string]any{"count_total": 0, "count": make([]int, 14), "percent": make([]float64, 14)}))
		return
	}
	rows, err := a.db.Query(r.Context(), "select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0 and (substat_all & $1) = $1", target)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	defer rows.Close()
	counts := make([]int, 15)
	for rows.Next() {
		var log EchoLog
		if err := rows.Scan(&log.Substat1, &log.Substat2, &log.Substat3, &log.Substat4, &log.Substat5); err != nil {
			writeJSON(w, appError("failed to get echo logs", 500))
			return
		}
		switch {
		case payload.Substat1 == 0:
			incrementPredictCount(counts, log.Substat1&substatMask)
		case payload.Substat1&substatMask != log.Substat1&substatMask:
			continue
		case payload.Substat2 == 0:
			incrementPredictCount(counts, log.Substat2&substatMask)
		case payload.Substat2&substatMask != log.Substat2&substatMask:
			continue
		case payload.Substat3 == 0:
			incrementPredictCount(counts, log.Substat3&substatMask)
		case payload.Substat3&substatMask != log.Substat3&substatMask:
			continue
		case payload.Substat4 == 0:
			incrementPredictCount(counts, log.Substat4&substatMask)
		case payload.Substat4&substatMask != log.Substat4&substatMask:
			continue
		case payload.Substat5 == 0:
			incrementPredictCount(counts, log.Substat5&substatMask)
		}
	}
	counts = counts[:13]
	total := 0
	for _, count := range counts {
		total += count
	}
	percent := make([]float64, len(counts))
	for i, count := range counts {
		if total > 0 {
			percent[i] = rounded(float64(count)/float64(total)*100, 1)
		}
	}
	writeJSON(w, success("predict echo substat", map[string]any{"count_total": total, "count": counts, "percent": percent}))
}
