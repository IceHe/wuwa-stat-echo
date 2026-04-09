package goapp

import (
	"net/http"
	"time"
)

func applyEchoChanges(existing *EchoLog, payload EchoLog) {
	if payload.Substat1 != 0 {
		existing.Substat1 = payload.Substat1
	}
	if payload.Substat2 != 0 {
		existing.Substat2 = payload.Substat2
	}
	if payload.Substat3 != 0 {
		existing.Substat3 = payload.Substat3
	}
	if payload.Substat4 != 0 {
		existing.Substat4 = payload.Substat4
	}
	if payload.Substat5 != 0 {
		existing.Substat5 = payload.Substat5
	}
	if payload.SubstatAll != 0 {
		existing.SubstatAll = payload.SubstatAll
	}
	if payload.S1Desc != "" {
		existing.S1Desc = payload.S1Desc
	}
	if payload.S2Desc != "" {
		existing.S2Desc = payload.S2Desc
	}
	if payload.S3Desc != "" {
		existing.S3Desc = payload.S3Desc
	}
	if payload.S4Desc != "" {
		existing.S4Desc = payload.S4Desc
	}
	if payload.S5Desc != "" {
		existing.S5Desc = payload.S5Desc
	}
	if payload.Clazz != "" {
		existing.Clazz = payload.Clazz
	}
	if payload.UserID != 0 {
		existing.UserID = payload.UserID
	}
	now := time.Now()
	existing.UpdatedAt = &now
}

func replaceEchoChanges(existing *EchoLog, payload EchoLog) {
	existing.Substat1 = payload.Substat1
	existing.Substat2 = payload.Substat2
	existing.Substat3 = payload.Substat3
	existing.Substat4 = payload.Substat4
	existing.Substat5 = payload.Substat5
	existing.SubstatAll = payload.SubstatAll
	existing.S1Desc = payload.S1Desc
	existing.S2Desc = payload.S2Desc
	existing.S3Desc = payload.S3Desc
	existing.S4Desc = payload.S4Desc
	existing.S5Desc = payload.S5Desc
	existing.Clazz = payload.Clazz
	existing.UserID = payload.UserID
	now := time.Now()
	existing.UpdatedAt = &now
}

func (a *App) handleCreateEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	now := time.Now()
	if payload.TunedAt == nil {
		payload.TunedAt = &now
	}
	payload.CreatedAt = &now
	payload.UpdatedAt = &now
	payload.OperatorID = operatorID
	row := a.db.QueryRow(r.Context(), "insert into wuwa_echo_log (substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, tuned_at, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", payload.Substat1, payload.Substat2, payload.Substat3, payload.Substat4, payload.Substat5, payload.SubstatAll, payload.S1Desc, payload.S2Desc, payload.S3Desc, payload.S4Desc, payload.S5Desc, payload.Clazz, payload.UserID, *operatorID, payload.TunedAt, payload.CreatedAt, payload.UpdatedAt)
	created, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), a.db, nil, created); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), a.db, nil, created); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	a.ws.send(*operatorID, map[string]any{"type": "create_echo_log", "data": created})
	writeJSON(w, success("create echo log", created))
}

func (a *App) handleUpdateEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	existing, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", payload.ID))
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	if existing.OperatorID == nil || (*existing.OperatorID != *operatorID && !canManage(r.Context())) {
		writeJSON(w, appError("not authorized to update this echo log", 403))
		return
	}
	beforeEcho := *existing
	replaceEchoChanges(existing, payload)
	row := a.db.QueryRow(r.Context(), "update wuwa_echo_log set substat1=$1, substat2=$2, substat3=$3, substat4=$4, substat5=$5, substat_all=$6, s1_desc=$7, s2_desc=$8, s3_desc=$9, s4_desc=$10, s5_desc=$11, clazz=$12, user_id=$13, updated_at=$14 where id=$15 returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", existing.Substat1, existing.Substat2, existing.Substat3, existing.Substat4, existing.Substat5, existing.SubstatAll, existing.S1Desc, existing.S2Desc, existing.S3Desc, existing.S4Desc, existing.S5Desc, existing.Clazz, existing.UserID, existing.UpdatedAt, existing.ID)
	updated, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), a.db, &beforeEcho, updated); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), a.db, &beforeEcho, updated); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	a.ws.send(*updated.OperatorID, map[string]any{"type": "update_echo_log", "data": updated})
	writeJSON(w, success("update echo log", updated))
}

func (a *App) handleTuneEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoTuneRequest
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())

	var echoLog *EchoLog
	now := time.Now()
	if payload.ID > 0 {
		row := tx.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", payload.ID)
		echoLog, err = a.scanEchoLog(row)
		if err != nil {
			writeJSON(w, appError("echo log not found", 404))
			return
		}
		if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
			writeJSON(w, appError("not authorized to tune this echo log", 403))
			return
		}
	} else {
		if payload.UserID == 0 {
			writeJSON(w, appError("user_id is required", 400))
			return
		}
		if payload.Clazz == "" {
			writeJSON(w, appError("clazz is required", 400))
			return
		}
		row := tx.QueryRow(r.Context(), "insert into wuwa_echo_log (user_id, clazz, tuned_at, created_at, updated_at, operator_id) values ($1,$2,$3,$4,$5,$6) returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", payload.UserID, payload.Clazz, now, now, now, *operatorID)
		echoLog, err = a.scanEchoLog(row)
		if err != nil {
			writeJSON(w, appError("failed to tune echo log", 500))
			return
		}
	}
	beforeEcho := *echoLog
	applyEchoChanges(echoLog, EchoLog{Substat1: payload.Substat1, Substat2: payload.Substat2, Substat3: payload.Substat3, Substat4: payload.Substat4, Substat5: payload.Substat5, SubstatAll: payload.SubstatAll, S1Desc: payload.S1Desc, S2Desc: payload.S2Desc, S3Desc: payload.S3Desc, S4Desc: payload.S4Desc, S5Desc: payload.S5Desc, Clazz: payload.Clazz, UserID: payload.UserID})
	row := tx.QueryRow(r.Context(), "update wuwa_echo_log set substat1=$1, substat2=$2, substat3=$3, substat4=$4, substat5=$5, substat_all=$6, s1_desc=$7, s2_desc=$8, s3_desc=$9, s4_desc=$10, s5_desc=$11, clazz=$12, user_id=$13, updated_at=$14 where id=$15 returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", echoLog.Substat1, echoLog.Substat2, echoLog.Substat3, echoLog.Substat4, echoLog.Substat5, echoLog.SubstatAll, echoLog.S1Desc, echoLog.S2Desc, echoLog.S3Desc, echoLog.S4Desc, echoLog.S5Desc, echoLog.Clazz, echoLog.UserID, echoLog.UpdatedAt, echoLog.ID)
	echoLog, err = a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	var tuneLog SubstatLog
	err = tx.QueryRow(r.Context(), "insert into wuwa_tune_log (user_id, echo_id, position, substat, value, operator_id) values ($1,$2,$3,$4,$5,$6) returning id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted", echoLog.UserID, echoLog.ID, payload.Position, payload.Substat, payload.Value, *operatorID).Scan(&tuneLog.ID, &tuneLog.Substat, &tuneLog.Value, &tuneLog.Position, &tuneLog.EchoID, &tuneLog.UserID, &tuneLog.OperatorID, &tuneLog.Timestamp, &tuneLog.Deleted)
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{tuneLog}, 1); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, echoLog); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, echoLog); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	a.ws.send(*operatorID, map[string]any{"type": "tune_echo_log", "data": map[string]any{"echo_log": echoLog, "tune_log": tuneLog}})
	writeJSON(w, success("tune echo log", map[string]any{"echo_log": echoLog, "tune_log": tuneLog}))
}
