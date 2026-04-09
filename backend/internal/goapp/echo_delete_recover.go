package goapp

import "net/http"

func (a *App) handleDeleteEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id))
	if err != nil {
		writeJSON(w, appError("echo log not found", 500))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete this echo log", 403))
		return
	}
	emptyEcho := echoLog.Substat1 == 0 && echoLog.Substat2 == 0 && echoLog.Substat3 == 0 && echoLog.Substat4 == 0 && echoLog.Substat5 == 0
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and deleted = 0", id)
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	result := map[string]any{}
	if emptyEcho {
		if _, err := tx.Exec(r.Context(), "delete from wuwa_tune_log where echo_id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if _, err := tx.Exec(r.Context(), "delete from wuwa_echo_log where id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		result = map[string]any{"deleted": "hard", "id": id}
	} else {
		beforeEcho := *echoLog
		afterEcho := beforeEcho
		afterEcho.Deleted = 1
		tag1, err := tx.Exec(r.Context(), "update wuwa_echo_log set deleted = 1 where id = $1", id)
		if err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set deleted = 1 where echo_id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		result = map[string]any{"rows_affected": tag1.RowsAffected()}
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, -1); err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	a.ws.send(*echoLog.OperatorID, map[string]any{"type": "delete_echo_log", "data": map[string]any{"id": id, "deleted": map[bool]string{true: "hard", false: "soft"}[emptyEcho]}})
	writeJSON(w, success("delete echo log", result))
}

func (a *App) handleRecoverEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id))
	if err != nil {
		writeJSON(w, appError("echo log not found", 500))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to recover this echo log", 403))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and deleted = 1", id)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	beforeEcho := *echoLog
	afterEcho := beforeEcho
	afterEcho.Deleted = 0
	tag, err := tx.Exec(r.Context(), "update wuwa_echo_log set deleted = 0 where id = $1", id)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set deleted = 0 where echo_id = $1", id); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, 1); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("recover echo log", map[string]int64{"rows_affected": tag.RowsAffected()}))
}

func (a *App) handleDeleteSubstatByEchoPos(w http.ResponseWriter, r *http.Request) {
	echoID := parseInt64Default(r.PathValue("echoId"), 0)
	pos := parseIntDefault(r.PathValue("pos"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", echoID))
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete substats for this echo log", 403))
		return
	}
	query := "update wuwa_tune_log set deleted = 1 where echo_id = $1 and position = $2"
	args := []any{echoID, pos}
	if !isManager {
		query += " and operator_id = $3"
		args = append(args, *operatorID)
	}
	selectQuery := "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and position = $2 and deleted = 0"
	selectArgs := []any{echoID, pos}
	if !isManager {
		selectQuery += " and operator_id = $3"
		selectArgs = append(selectArgs, *operatorID)
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), selectQuery, selectArgs...)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	tag, err := tx.Exec(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, -1); err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("delete substat log", map[string]int64{"rows_affected": tag.RowsAffected()}))
}
