package goapp

import (
	"context"
	"net/http"
	"strconv"
)

type EchoDecisionRequest struct {
	Echo       EchoLog `json:"echo"`
	UserID     int64   `json:"user_id,omitempty"`
	Resonator  string  `json:"resonator"`
	Cost       string  `json:"cost"`
	Goal       string  `json:"goal"`
	TargetBits int64   `json:"target_bits,omitempty"`
	Trials     int     `json:"trials,omitempty"`
	Window     string  `json:"window,omitempty"`
}

type simulatorSample struct {
	Substats [5]int64
}

type simulationSummary struct {
	Trials            int              `json:"trials"`
	HitProb           float64          `json:"hit_prob"`
	HighRollProb      float64          `json:"high_roll_prob"`
	ExpectedScore     float64          `json:"expected_score"`
	ExpectedTunerCost int              `json:"expected_tuner_cost"`
	ExpectedExpCost   int              `json:"expected_exp_cost"`
	ResultBuckets     []map[string]any `json:"result_buckets"`
}

func (a *App) handleDecisionEchoNextStep(w http.ResponseWriter, r *http.Request) {
	req, err := a.readDecisionRequest(r)
	if err != nil {
		writeJSON(w, appError(err.Error(), 400))
		return
	}
	samples, err := a.loadSimulatorSamples(r.Context(), req.UserID, req.Window)
	if err != nil {
		writeJSON(w, appError("failed to load echo samples", 500))
		return
	}

	currentScore := scoreEcho(req.Echo, req.Resonator, req.Cost).SubstatAll
	maxScore := maxEchoScore(req.Resonator, req.Cost)
	percentile := scorePercentile(samples, req.Echo, req.Resonator, req.Cost)
	targetBits := req.TargetBits

	nextProb := a.computeNextGoodRollProbability(samples, req.Echo, req.Resonator, targetBits)
	finishSummary := a.simulateEchoFuture(samples, req, -1)
	lockedValue := 0.0
	if maxScore > 0 {
		lockedValue = rounded(currentScore/maxScore, 4)
	}
	resp := map[string]any{
		"current_score":           currentScore,
		"percentile":              percentile,
		"effective_substat_count": countEffectiveSubstats(req.Echo, req.Resonator),
		"locked_value":            lockedValue,
		"continue_to_next_prob":   nextProb,
		"continue_to_finish_prob": finishSummary.HitProb,
		"expected_extra_tuner":    finishSummary.ExpectedTunerCost,
		"expected_extra_exp":      finishSummary.ExpectedExpCost,
		"recommendation":          decisionRecommendation(nextProb, finishSummary.HitProb, lockedValue, percentile),
		"reasons":                 buildDecisionReasons(req.Echo, targetBits, nextProb, finishSummary.HitProb, percentile),
	}
	writeJSON(w, success("decision echo next step", resp))
}

func (a *App) handleSimulatorEchoFuture(w http.ResponseWriter, r *http.Request) {
	req, err := a.readDecisionRequest(r)
	if err != nil {
		writeJSON(w, appError(err.Error(), 400))
		return
	}
	samples, err := a.loadSimulatorSamples(r.Context(), req.UserID, req.Window)
	if err != nil {
		writeJSON(w, appError("failed to load echo samples", 500))
		return
	}
	writeJSON(w, success("simulator echo future", a.simulateEchoFuture(samples, req, -1)))
}

func (a *App) handleSimulatorEchoCompare(w http.ResponseWriter, r *http.Request) {
	req, err := a.readDecisionRequest(r)
	if err != nil {
		writeJSON(w, appError(err.Error(), 400))
		return
	}
	samples, err := a.loadSimulatorSamples(r.Context(), req.UserID, req.Window)
	if err != nil {
		writeJSON(w, appError("failed to load echo samples", 500))
		return
	}
	currentScore := scoreEcho(req.Echo, req.Resonator, req.Cost).SubstatAll
	targetBitsHit := (req.Echo.SubstatAll & req.TargetBits) == req.TargetBits
	stopHit := targetBitsHit && currentScore >= goalThreshold(req.Goal, maxEchoScore(req.Resonator, req.Cost))
	resp := map[string]any{
		"current_score": currentScore,
		"goal":          req.Goal,
		"strategies": []map[string]any{
			{
				"strategy": "stop_now",
				"summary": simulationSummary{
					Trials:            1,
					HitProb:           rounded(boolRate(stopHit), 4),
					HighRollProb:      rounded(boolRate(scoreBucket(currentScore, maxEchoScore(req.Resonator, req.Cost)) == "神品"), 4),
					ExpectedScore:     currentScore,
					ExpectedTunerCost: 0,
					ExpectedExpCost:   0,
					ResultBuckets:     []map[string]any{{"label": scoreBucket(currentScore, maxEchoScore(req.Resonator, req.Cost)), "rate": 1}},
				},
			},
			{
				"strategy": "continue_once",
				"summary":  a.simulateEchoFuture(samples, req, 1),
			},
			{
				"strategy": "continue_to_end",
				"summary":  a.simulateEchoFuture(samples, req, -1),
			},
		},
	}
	writeJSON(w, success("simulator echo compare", resp))
}

func (a *App) readDecisionRequest(r *http.Request) (*EchoDecisionRequest, error) {
	var req EchoDecisionRequest
	if err := readJSON(r, &req); err != nil {
		return nil, err
	}
	req.Echo = normalizeDecisionEcho(req.Echo)
	if req.UserID <= 0 {
		req.UserID = req.Echo.UserID
	}
	if req.Cost == "" {
		req.Cost = "1C"
	}
	if req.Goal == "" {
		req.Goal = "毕业"
	}
	if req.TargetBits == 0 {
		req.TargetBits = defaultTargetBits(req.Resonator, req.Goal)
	}
	if req.Trials <= 0 {
		req.Trials = 5000
	}
	if req.Trials > 50000 {
		req.Trials = 50000
	}
	return &req, nil
}

func normalizeDecisionEcho(e EchoLog) EchoLog {
	e.SubstatAll = (e.Substat1 | e.Substat2 | e.Substat3 | e.Substat4 | e.Substat5) & substatMask
	return e
}

func (a *App) loadSimulatorSamples(ctx context.Context, userID int64, rawWindow string) ([]simulatorSample, error) {
	window := parseStatsWindow(rawWindow)
	query := "select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	args := []any{}
	arg := 1
	if userID > 0 {
		query += " and user_id = $" + strconv.Itoa(arg)
		args = append(args, userID)
		arg++
	}
	if since := window.sinceTime(); since != nil {
		query += " and updated_at >= $" + strconv.Itoa(arg)
		args = append(args, *since)
		arg++
	}
	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	samples := make([]simulatorSample, 0, 256)
	for rows.Next() {
		var item simulatorSample
		if err := rows.Scan(&item.Substats[0], &item.Substats[1], &item.Substats[2], &item.Substats[3], &item.Substats[4]); err != nil {
			return nil, err
		}
		samples = append(samples, item)
	}
	return samples, rows.Err()
}
