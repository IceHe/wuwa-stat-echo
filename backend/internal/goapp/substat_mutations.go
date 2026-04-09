package goapp

import (
	"fmt"
	"net/http"
)

func (a *App) handleDeleteTuneLogByID(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	var tuneLog SubstatLog
	err := a.db.QueryRow(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where id = $1", id).Scan(&tuneLog.ID, &tuneLog.Substat, &tuneLog.Value, &tuneLog.Position, &tuneLog.EchoID, &tuneLog.UserID, &tuneLog.OperatorID, &tuneLog.Timestamp, &tuneLog.Deleted)
	if err != nil {
		writeJSON(w, appError("tune log not found", 500))
		return
	}
	if tuneLog.OperatorID == nil || (*tuneLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete this tune log", 403))
		return
	}
	var echoOperatorID *int64
	_ = a.db.QueryRow(r.Context(), "select operator_id from wuwa_echo_log where id = $1", tuneLog.EchoID).Scan(&echoOperatorID)
	if echoOperatorID != nil && *echoOperatorID != *operatorID && !isManager {
		writeJSON(w, appError("not authorized to delete tune log for this echo log", 403))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	defer tx.Rollback(r.Context())

	if tuneLog.Deleted == 0 {
		if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{tuneLog}, -1); err != nil {
			writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
			return
		}
	}
	tag, err := tx.Exec(r.Context(), "delete from wuwa_tune_log where id = $1", id)
	if err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success(fmt.Sprintf("delete tune log %d", id), map[string]int64{"row_deleted": tag.RowsAffected()}))
}

func (a *App) handleAddTuneLog(w http.ResponseWriter, r *http.Request) {
	var payload SubstatLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	defer tx.Rollback(r.Context())

	var created SubstatLog
	err = tx.QueryRow(r.Context(), "insert into wuwa_tune_log (user_id, echo_id, position, substat, value, operator_id) values ($1, $2, $3, $4, $5, $6) returning id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted", payload.UserID, payload.EchoID, payload.Position, payload.Substat, payload.Value, *operatorID).Scan(&created.ID, &created.Substat, &created.Value, &created.Position, &created.EchoID, &created.UserID, &created.OperatorID, &created.Timestamp, &created.Deleted)
	if err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{created}, 1); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("add tune log", map[string]any{}))
}

func (a *App) handleWriteEchoSubstatAll(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all from wuwa_echo_log where deleted = 0")
	if err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	defer rows.Close()
	var total, successTotal int64
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	defer tx.Rollback(r.Context())
	for rows.Next() {
		total++
		var id, s1, s2, s3, s4, s5, substatAll int64
		if err := rows.Scan(&id, &s1, &s2, &s3, &s4, &s5, &substatAll); err != nil {
			writeJSON(w, appError("failed to write substat all", 500))
			return
		}
		if substatAll == 0 {
			calculated := (s1 | s2 | s3 | s4 | s5) & substatMask
			if _, err := tx.Exec(r.Context(), "update wuwa_echo_log set substat_all = $1 where id = $2", calculated, id); err != nil {
				writeJSON(w, appError("failed to write substat all", 500))
				return
			}
			successTotal++
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	writeJSON(w, success("write substat all", map[string]int64{"success_total": successTotal, "total": total}))
}

func (a *App) handleWriteSubstatUserID(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "select id, echo_id, user_id from wuwa_tune_log where deleted = 0 order by id desc")
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	defer rows.Close()
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	defer tx.Rollback(r.Context())
	var total, successTotal int64
	for rows.Next() {
		total++
		var id, echoID, userID int64
		if err := rows.Scan(&id, &echoID, &userID); err != nil {
			writeJSON(w, appError("failed to get stats", 500))
			return
		}
		if echoID > 0 && userID == 0 {
			var echoUserID int64
			if err := tx.QueryRow(r.Context(), "select user_id from wuwa_echo_log where id = $1", echoID).Scan(&echoUserID); err == nil && echoUserID > 0 {
				if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set user_id = $1 where id = $2", echoUserID, id); err != nil {
					writeJSON(w, appError("failed to get stats", 500))
					return
				}
				successTotal++
			}
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	writeJSON(w, success("write id", map[string]int64{"success_total": successTotal, "total": total}))
}
