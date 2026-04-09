package goapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (a *App) handleListEchoLogs(w http.ResponseWriter, r *http.Request) {
	pageNum := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	offset := (pageNum - 1) * pageSize
	rows, err := a.db.Query(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log order by updated_at desc offset $1 limit $2", offset, pageSize)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	var total int64
	if err := a.db.QueryRow(r.Context(), "select count(id) from wuwa_echo_log where deleted = 0").Scan(&total); err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	writeJSON(w, page("echo logs", items, total, pageNum, pageSize))
}

func (a *App) handleGetEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	var row pgx.Row
	if id > 0 {
		row = a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id)
	} else {
		row = a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0 and operator_id = $1 order by updated_at desc limit 1", *operatorID)
	}
	echoLog, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	writeJSON(w, success("echo log", map[string]any{
		"id":          echoLog.ID,
		"substat1":    echoLog.Substat1,
		"substat2":    echoLog.Substat2,
		"substat3":    echoLog.Substat3,
		"substat4":    echoLog.Substat4,
		"substat5":    echoLog.Substat5,
		"substat_all": echoLog.SubstatAll,
		"s1_desc":     echoLog.S1Desc,
		"s2_desc":     echoLog.S2Desc,
		"s3_desc":     echoLog.S3Desc,
		"s4_desc":     echoLog.S4Desc,
		"s5_desc":     echoLog.S5Desc,
		"clazz":       echoLog.Clazz,
		"user_id":     echoLog.UserID,
		"operator_id": echoLog.OperatorID,
		"deleted":     echoLog.Deleted,
		"tuned_at":    echoLog.TunedAt,
		"created_at":  echoLog.CreatedAt,
		"updated_at":  echoLog.UpdatedAt,
		"pos_total":   a.posTotalExcludingEcho(*echoLog, a.getCachedTuneStats()),
	}))
}

func (a *App) handleFindEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoFindRequest
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	hasSubstatFilter := (payload.Substat1 | payload.Substat2 | payload.Substat3 | payload.Substat4 | payload.Substat5) != 0
	keyword := strings.TrimSpace(payload.Keyword)
	if !hasSubstatFilter && payload.ID <= 0 && payload.UserID <= 0 && payload.Clazz == "" && keyword == "" {
		writeJSON(w, success("no search condition specified, return empty list", []EchoLog{}))
		return
	}
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0"
	var args []any
	arg := 1
	if payload.ID > 0 {
		query += fmt.Sprintf(" and id = $%d", arg)
		args = append(args, payload.ID)
		arg++
	}
	for _, filter := range []struct {
		column string
		bits   int64
	}{
		{"substat1", payload.Substat1},
		{"substat2", payload.Substat2},
		{"substat3", payload.Substat3},
		{"substat4", payload.Substat4},
		{"substat5", payload.Substat5},
	} {
		if filter.bits == 0 {
			continue
		}
		if filter.bits&^int64(substatMask) == 0 {
			query += fmt.Sprintf(" and (%s & $%d) = $%d", filter.column, arg, arg)
		} else {
			query += fmt.Sprintf(" and %s = $%d", filter.column, arg)
		}
		args = append(args, filter.bits)
		arg++
	}
	if payload.UserID > 0 {
		query += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, payload.UserID)
		arg++
	}
	if payload.Clazz != "" {
		query += fmt.Sprintf(" and clazz = $%d", arg)
		args = append(args, payload.Clazz)
		arg++
	}
	if keyword != "" {
		query += fmt.Sprintf(" and (clazz ilike $%d or s1_desc ilike $%d or s2_desc ilike $%d or s3_desc ilike $%d or s4_desc ilike $%d or s5_desc ilike $%d or cast(user_id as text) ilike $%d or cast(id as text) ilike $%d)", arg, arg, arg, arg, arg, arg, arg, arg)
		args = append(args, "%"+keyword+"%")
		arg++
	}
	query += fmt.Sprintf(" order by updated_at desc limit %d", pageSize)
	rows, err := a.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	writeJSON(w, success("find echo logs", items))
}

func (a *App) handleEchoLogsAnalysis(w http.ResponseWriter, r *http.Request) {
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	targetBits := parseInt64Default(r.URL.Query().Get("target_bits"), 0b11)
	substatSinceDate := strings.TrimSpace(r.URL.Query().Get("substat_since_date"))
	window := parseStatsWindow(r.URL.Query().Get("window"))
	if window.isAll() && size == 0 && substatSinceDate == "" {
		if data, err := a.loadEchoSummaryFromAggregate(r.Context(), userID, targetBits); err == nil && data != nil {
			if userID > 0 {
				if globalData, globalErr := a.loadEchoSummaryFromAggregate(r.Context(), 0, targetBits); globalErr == nil && globalData != nil {
					data["baseline_compare"] = map[string]any{
						"target_rate": buildRateComparison(
							data["target_rate_stats"].(*ProportionStat),
							globalData["target_rate_stats"].(*ProportionStat),
						),
					}
				}
			}
			data["window"] = window.Name
			writeJSON(w, success("echo logs analysis", data))
			return
		}
	}
	effectiveSize := window.applyLimit(size)
	items, total, err := a.loadEchoLogsAnalysisItems(r.Context(), userID, effectiveSize, targetBits, window, substatSinceDate)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	resp := computeEchoLogsAnalysisFromItems(items, total, targetBits)
	resp["window"] = window.Name
	if userID > 0 {
		globalItems, globalTotal, globalErr := a.loadEchoLogsAnalysisItems(r.Context(), 0, effectiveSize, targetBits, window, substatSinceDate)
		if globalErr != nil {
			writeJSON(w, appError("failed to get echo logs", 500))
			return
		}
		globalResp := computeEchoLogsAnalysisFromItems(globalItems, globalTotal, targetBits)
		resp["baseline_compare"] = map[string]any{
			"target_rate": buildRateComparison(resp["target_rate_stats"].(*ProportionStat), globalResp["target_rate_stats"].(*ProportionStat)),
		}
	}
	writeJSON(w, success("echo logs analysis", resp))
}

func (a *App) handleViewerScoreTemplateSync(w http.ResponseWriter, r *http.Request) {
	var payload ScoreTemplateSyncRequest
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to sync score template", 400))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	field := strings.TrimSpace(payload.Field)
	value := strings.TrimSpace(payload.Value)
	if field == "" || value == "" {
		writeJSON(w, appError("field and value are required", 400))
		return
	}
	switch field {
	case "resonator", "cost":
	default:
		writeJSON(w, appError("invalid field", 400))
		return
	}
	a.ws.send(*operatorID, map[string]any{
		"type": "score_template_changed",
		"data": map[string]any{
			"field":     field,
			"value":     value,
			"resonator": strings.TrimSpace(payload.Resonator),
			"cost":      strings.TrimSpace(payload.Cost),
		},
	})
	writeJSON(w, success("viewer score template synced", map[string]any{}))
}

func (a *App) loadEchoLogsAnalysisItems(ctx context.Context, userID int64, effectiveSize int, targetBits int64, window statsWindow, substatSinceDate string) ([]EchoLog, int64, error) {
	selectSQL := "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0"
	countSQL := "select count(id) from wuwa_echo_log where deleted = 0"
	var args []any
	arg := 1
	if userID > 0 {
		selectSQL += fmt.Sprintf(" and user_id = $%d", arg)
		countSQL += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, userID)
		arg++
		if substatSinceDate != "" {
			parsed, err := time.Parse("2006-01-02", substatSinceDate)
			if err == nil {
				startAt := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 4, 0, 0, 0, parsed.Location())
				rows, err := a.db.Query(ctx, "select echo_id from wuwa_tune_log where deleted = 0 and user_id = $1 and timestamp >= $2", userID, startAt)
				if err == nil {
					defer rows.Close()
					idsMap := map[int64]struct{}{}
					var ids []int64
					for rows.Next() {
						var echoID int64
						if rows.Scan(&echoID) == nil {
							if _, ok := idsMap[echoID]; !ok {
								idsMap[echoID] = struct{}{}
								ids = append(ids, echoID)
							}
						}
					}
					if len(ids) > 0 {
						selectSQL += fmt.Sprintf(" and id = any($%d)", arg)
						countSQL += fmt.Sprintf(" and id = any($%d)", arg)
						args = append(args, ids)
						arg++
					} else {
						selectSQL += " and id = -1"
						countSQL += " and id = -1"
					}
				}
			}
		}
	}
	if since := window.sinceTime(); since != nil {
		selectSQL += fmt.Sprintf(" and updated_at >= $%d", arg)
		countSQL += fmt.Sprintf(" and updated_at >= $%d", arg)
		args = append(args, *since)
		arg++
	}
	selectSQL += " order by updated_at desc"
	if effectiveSize > 0 {
		selectSQL += fmt.Sprintf(" limit %d", effectiveSize)
	}
	rows, err := a.db.Query(ctx, selectSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := a.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if effectiveSize > 0 {
		total = int64(len(items))
	}
	return items, total, nil
}
