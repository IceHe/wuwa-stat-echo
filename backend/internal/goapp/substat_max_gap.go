package goapp

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const substatMaxGapForceCooldown = time.Hour

type substatGapState struct {
	userID    int64
	seen      bool
	lastIndex int
	lastID    int64
	count     int
	leading   int
	trailing  int
	maxGap    int
	startID   int64
	endID     int64
}

type maxGapEvent struct {
	id     int64
	userID int64
	hit    bool
}

func cloneSubstatMaxGapResponse(src *SubstatMaxGapResponse) *SubstatMaxGapResponse {
	if src == nil {
		return nil
	}
	out := *src
	out.Rows = append([]SubstatMaxGapRow(nil), src.Rows...)
	if src.DoubleCritEchoGap != nil {
		doubleCritEchoGap := *src.DoubleCritEchoGap
		out.DoubleCritEchoGap = &doubleCritEchoGap
	}
	return &out
}

func (a *App) getCachedSubstatMaxGap(userID int64) *SubstatMaxGapResponse {
	a.maxGapMu.RLock()
	defer a.maxGapMu.RUnlock()
	return cloneSubstatMaxGapResponse(a.substatMaxGapCache[userID])
}

func (a *App) storeCachedSubstatMaxGap(userID int64, data *SubstatMaxGapResponse) {
	a.maxGapMu.Lock()
	defer a.maxGapMu.Unlock()
	a.substatMaxGapCache[userID] = cloneSubstatMaxGapResponse(data)
}

func (a *App) handleSubstatMaxGap(w http.ResponseWriter, r *http.Request) {
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	force := parseIntDefault(r.URL.Query().Get("force"), 0) == 1

	resp, err := a.loadSubstatMaxGap(r.Context(), userID, force)
	if err != nil {
		writeJSON(w, appError("failed to get substat max gap", 500))
		return
	}
	writeJSON(w, success("substat max gap", resp))
}

func (a *App) loadSubstatMaxGap(ctx context.Context, userID int64, force bool) (*SubstatMaxGapResponse, error) {
	now := time.Now()
	cached := a.getCachedSubstatMaxGap(userID)
	if cached == nil {
		computed, err := a.computeSubstatMaxGap(ctx, userID)
		if err != nil {
			return nil, err
		}
		if force {
			computed.LastForcedRefreshAt = &now
			nextAllowed := now.Add(substatMaxGapForceCooldown)
			computed.RefreshAvailableAt = &nextAllowed
		}
		a.storeCachedSubstatMaxGap(userID, computed)
		resp := cloneSubstatMaxGapResponse(computed)
		resp.CacheHit = false
		resp.ForceApplied = force
		return resp, nil
	}

	if !force {
		resp := cloneSubstatMaxGapResponse(cached)
		resp.CacheHit = true
		resp.ForceApplied = false
		resp.RefreshBlocked = false
		return resp, nil
	}

	if cached.LastForcedRefreshAt != nil {
		nextAllowed := cached.LastForcedRefreshAt.Add(substatMaxGapForceCooldown)
		if now.Before(nextAllowed) {
			resp := cloneSubstatMaxGapResponse(cached)
			resp.CacheHit = true
			resp.ForceApplied = false
			resp.RefreshBlocked = true
			resp.RefreshAvailableAt = &nextAllowed
			return resp, nil
		}
	}

	computed, err := a.computeSubstatMaxGap(ctx, userID)
	if err != nil {
		return nil, err
	}
	computed.LastForcedRefreshAt = &now
	nextAllowed := now.Add(substatMaxGapForceCooldown)
	computed.RefreshAvailableAt = &nextAllowed
	a.storeCachedSubstatMaxGap(userID, computed)
	resp := cloneSubstatMaxGapResponse(computed)
	resp.CacheHit = false
	resp.ForceApplied = true
	resp.RefreshBlocked = false
	return resp, nil
}

func (a *App) computeSubstatMaxGap(ctx context.Context, userID int64) (*SubstatMaxGapResponse, error) {
	query := "select id, substat, user_id from wuwa_tune_log where deleted = 0"
	args := []any{}
	if userID > 0 {
		query += " and user_id = $1"
		args = append(args, userID)
	}
	query += " order by user_id asc, id asc"

	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bestStates := make([]substatGapState, len(substatDefs))
	perUserStates := map[int64][]substatGapState{}
	userTotals := map[int64]int{}
	total := 0
	for rows.Next() {
		var id int64
		var substat int
		var rowUserID int64
		if err := rows.Scan(&id, &substat, &rowUserID); err != nil {
			return nil, err
		}
		if substat < 0 || substat >= len(bestStates) {
			total++
			continue
		}
		if _, ok := perUserStates[rowUserID]; !ok {
			perUserStates[rowUserID] = make([]substatGapState, len(substatDefs))
			for index := range perUserStates[rowUserID] {
				perUserStates[rowUserID][index].userID = rowUserID
			}
		}
		userIndex := userTotals[rowUserID]
		state := &perUserStates[rowUserID][substat]
		state.count++
		if !state.seen {
			state.seen = true
			state.leading = userIndex
		} else {
			gap := userIndex - state.lastIndex - 1
			if gap > state.maxGap {
				state.maxGap = gap
				state.startID = state.lastID
				state.endID = id
			}
		}
		state.lastIndex = userIndex
		state.lastID = id
		userTotals[rowUserID] = userIndex + 1
		total++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	doubleCritEchoGap, echoLogTotal, err := a.computeDoubleCritEchoMaxGap(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := &SubstatMaxGapResponse{
		UserID:            userID,
		ScopeLabel:        map[bool]string{true: "全部玩家", false: "玩家 " + strconv.FormatInt(userID, 10)}[userID == 0],
		TuneLogTotal:      total,
		EchoLogTotal:      echoLogTotal,
		DoubleCritEchoGap: doubleCritEchoGap,
		Rows:              make([]SubstatMaxGapRow, 0, len(substatDefs)),
	}
	now := time.Now()
	result.GeneratedAt = &now
	for ownerUserID, states := range perUserStates {
		userTotal := userTotals[ownerUserID]
		for index := range states {
			state := &states[index]
			if !state.seen {
				state.leading = userTotal
				state.trailing = userTotal
				state.maxGap = userTotal
				state.startID = -1
				state.endID = -1
				continue
			}
			state.trailing = userTotal - state.lastIndex - 1
			if state.leading >= state.maxGap {
				state.maxGap = state.leading
				state.startID = -1
				state.endID = state.lastID
			}
			if state.trailing > state.maxGap {
				state.maxGap = state.trailing
				state.startID = state.lastID
				state.endID = -1
			}
			best := &bestStates[index]
			if !best.seen || state.maxGap > best.maxGap {
				*best = *state
			}
		}
	}
	for _, def := range substatDefs {
		state := bestStates[def.Number]
		if !state.seen {
			state.leading = total
			state.trailing = total
			state.maxGap = total
			state.startID = -1
			state.endID = -1
		}
		result.Rows = append(result.Rows, SubstatMaxGapRow{
			Substat:         def.Number,
			Name:            def.Name,
			NameCN:          def.NameCN,
			OwnerUserID:     state.userID,
			MaxGap:          state.maxGap,
			OccurrenceCount: state.count,
			LeadingGap:      state.leading,
			TrailingGap:     state.trailing,
			MaxGapStartID:   state.startID,
			MaxGapEndID:     state.endID,
		})
	}
	return result, nil
}

func (a *App) computeDoubleCritEchoMaxGap(ctx context.Context, userID int64) (*MaxGapMetricRow, int, error) {
	query := "select id, user_id, substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	args := []any{}
	if userID > 0 {
		query += " and user_id = $1"
		args = append(args, userID)
	}
	query += " order by user_id asc, id asc"

	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := []maxGapEvent{}
	for rows.Next() {
		var id int64
		var rowUserID int64
		var substat1, substat2, substat3, substat4, substat5 int64
		if err := rows.Scan(&id, &rowUserID, &substat1, &substat2, &substat3, &substat4, &substat5); err != nil {
			return nil, 0, err
		}
		substatAll := (substat1 | substat2 | substat3 | substat4 | substat5) & substatMask
		events = append(events, maxGapEvent{
			id:     id,
			userID: rowUserID,
			hit:    substatAll&0b11 == 0b11,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	state, total := computeMaxGapFromEvents(events)
	return &MaxGapMetricRow{
		Label:           "双暴声骸",
		OwnerUserID:     state.userID,
		MaxGap:          state.maxGap,
		OccurrenceCount: state.count,
		LeadingGap:      state.leading,
		TrailingGap:     state.trailing,
		MaxGapStartID:   state.startID,
		MaxGapEndID:     state.endID,
	}, total, nil
}

func computeMaxGapFromEvents(events []maxGapEvent) (substatGapState, int) {
	best := substatGapState{startID: -1, endID: -1}
	perUserStates := map[int64]*substatGapState{}
	userTotals := map[int64]int{}

	for _, event := range events {
		if _, ok := perUserStates[event.userID]; !ok {
			perUserStates[event.userID] = &substatGapState{
				userID:  event.userID,
				startID: -1,
				endID:   -1,
			}
		}
		userIndex := userTotals[event.userID]
		if event.hit {
			state := perUserStates[event.userID]
			state.count++
			if !state.seen {
				state.seen = true
				state.leading = userIndex
			} else {
				gap := userIndex - state.lastIndex - 1
				if gap > state.maxGap {
					state.maxGap = gap
					state.startID = state.lastID
					state.endID = event.id
				}
			}
			state.lastIndex = userIndex
			state.lastID = event.id
		}
		userTotals[event.userID] = userIndex + 1
	}

	for ownerUserID, state := range perUserStates {
		userTotal := userTotals[ownerUserID]
		if !state.seen {
			state.leading = userTotal
			state.trailing = userTotal
			state.maxGap = userTotal
			state.startID = -1
			state.endID = -1
		} else {
			state.trailing = userTotal - state.lastIndex - 1
			if state.leading >= state.maxGap {
				state.maxGap = state.leading
				state.startID = -1
				state.endID = state.lastID
			}
			if state.trailing > state.maxGap {
				state.maxGap = state.trailing
				state.startID = state.lastID
				state.endID = -1
			}
		}
		if best.userID == 0 && !best.seen && best.maxGap == 0 && best.startID == -1 && best.endID == -1 {
			best = *state
			continue
		}
		if state.maxGap > best.maxGap {
			best = *state
		}
	}

	return best, len(events)
}
