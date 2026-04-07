package goapp

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
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
			map[string]any{
				"strategy": "continue_once",
				"summary":  a.simulateEchoFuture(samples, req, 1),
			},
			map[string]any{
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

func (a *App) computeNextGoodRollProbability(samples []simulatorSample, echo EchoLog, resonator string, targetBits int64) float64 {
	pos := filledSubstatSlots(echo)
	if pos >= 5 {
		return 0
	}
	dist := nextRollDistribution(samples, echo, pos)
	if len(dist) == 0 {
		return 0
	}
	total := 0
	good := 0
	for encoded, count := range dist {
		total += count
		mask := encoded & substatMask
		if (mask&targetBits) != 0 || substatWeight(encoded, resonator) >= 0.2 {
			good += count
		}
	}
	if total == 0 {
		return 0
	}
	return rounded(float64(good)/float64(total), 4)
}

func (a *App) simulateEchoFuture(samples []simulatorSample, req *EchoDecisionRequest, steps int) simulationSummary {
	trials := req.Trials
	if len(samples) == 0 {
		trials = 0
	}
	maxScore := maxEchoScore(req.Resonator, req.Cost)
	threshold := goalThreshold(req.Goal, maxScore)
	expCost, tunerCost := extraUpgradeCost(req.Echo, steps)
	if trials == 0 {
		return simulationSummary{
			Trials:            0,
			ExpectedTunerCost: tunerCost,
			ExpectedExpCost:   expCost,
			ResultBuckets:     []map[string]any{},
		}
	}

	buckets := map[string]int{}
	hitCount := 0
	highRollCount := 0
	scoreSum := 0.0
	for i := 0; i < trials; i++ {
		final := simulateEchoRollout(samples, req.Echo, steps)
		finalScore := scoreEcho(final, req.Resonator, req.Cost).SubstatAll
		scoreSum += finalScore
		bucket := scoreBucket(finalScore, maxScore)
		buckets[bucket]++
		if isGoalMet(final, finalScore, req.TargetBits, threshold) {
			hitCount++
		}
		if bucket == "神品" || bucket == "大毕业" {
			highRollCount++
		}
	}

	order := []string{"止步中等", "小毕业", "大毕业", "神品"}
	resultBuckets := make([]map[string]any, 0, len(order))
	for _, label := range order {
		if buckets[label] == 0 {
			continue
		}
		resultBuckets = append(resultBuckets, map[string]any{
			"label": label,
			"rate":  rounded(float64(buckets[label])/float64(trials), 4),
		})
	}
	return simulationSummary{
		Trials:            trials,
		HitProb:           rounded(float64(hitCount)/float64(trials), 4),
		HighRollProb:      rounded(float64(highRollCount)/float64(trials), 4),
		ExpectedScore:     rounded(scoreSum/float64(trials), 2),
		ExpectedTunerCost: tunerCost,
		ExpectedExpCost:   expCost,
		ResultBuckets:     resultBuckets,
	}
}

func simulateEchoRollout(samples []simulatorSample, base EchoLog, steps int) EchoLog {
	echo := normalizeDecisionEcho(base)
	remaining := 5 - filledSubstatSlots(echo)
	if steps >= 0 && steps < remaining {
		remaining = steps
	}
	for step := 0; step < remaining; step++ {
		pos := filledSubstatSlots(echo)
		dist := nextRollDistribution(samples, echo, pos)
		picked, ok := pickEncodedSubstat(dist)
		if !ok {
			break
		}
		assignSubstatAt(&echo, pos, picked)
		echo.SubstatAll |= picked & substatMask
	}
	return echo
}

func nextRollDistribution(samples []simulatorSample, echo EchoLog, pos int) map[int64]int {
	current := currentSubstatSlice(echo)
	for _, mode := range []string{"exact", "mask", "global"} {
		dist := map[int64]int{}
		for _, sample := range samples {
			if pos < 0 || pos >= len(sample.Substats) || sample.Substats[pos] == 0 {
				continue
			}
			if (sample.Substats[pos] & echo.SubstatAll) != 0 {
				continue
			}
			if !sampleMatchesPrefix(sample.Substats[:], current, pos, mode) {
				continue
			}
			dist[sample.Substats[pos]]++
		}
		if len(dist) > 0 {
			return dist
		}
	}
	return nil
}

func sampleMatchesPrefix(sample []int64, current []int64, prefixLen int, mode string) bool {
	for i := 0; i < prefixLen; i++ {
		if current[i] == 0 {
			return false
		}
		switch mode {
		case "exact":
			if sample[i] != current[i] {
				return false
			}
		case "mask":
			if (sample[i] & substatMask) != (current[i] & substatMask) {
				return false
			}
		case "global":
			return true
		}
	}
	return true
}

func pickEncodedSubstat(dist map[int64]int) (int64, bool) {
	if len(dist) == 0 {
		return 0, false
	}
	keys := make([]int64, 0, len(dist))
	total := 0
	for key, count := range dist {
		if count <= 0 {
			continue
		}
		keys = append(keys, key)
		total += count
	}
	if total <= 0 {
		return 0, false
	}
	slices.Sort(keys)
	pick := rand.Intn(total)
	for _, key := range keys {
		pick -= dist[key]
		if pick < 0 {
			return key, true
		}
	}
	return keys[len(keys)-1], true
}

func currentSubstatSlice(e EchoLog) []int64 {
	return []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5}
}

func assignSubstatAt(e *EchoLog, pos int, value int64) {
	switch pos {
	case 0:
		e.Substat1 = value
	case 1:
		e.Substat2 = value
	case 2:
		e.Substat3 = value
	case 3:
		e.Substat4 = value
	case 4:
		e.Substat5 = value
	}
}

func filledSubstatSlots(e EchoLog) int {
	count := 0
	for _, substat := range []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5} {
		if substat != 0 {
			count++
		}
	}
	return count
}

func extraUpgradeCost(e EchoLog, steps int) (expCost int, tunerCost int) {
	remaining := 5 - filledSubstatSlots(e)
	if steps >= 0 && steps < remaining {
		remaining = steps
	}
	substatCount := filledSubstatSlots(e)
	expRaw := 0
	for i := 0; i < remaining; i++ {
		nextCount := substatCount + 1
		expRaw += expTable[substatCount][nextCount]
		substatCount = nextCount
	}
	return int(math.Ceil(float64(expRaw) / expGold)), remaining * 10
}

func countEffectiveSubstats(e EchoLog, resonator string) int {
	count := 0
	for _, substat := range []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5} {
		if substat != 0 && substatWeight(substat, resonator) >= 0.2 {
			count++
		}
	}
	return count
}

func substatWeight(encoded int64, resonator string) float64 {
	num := bitPos(encoded & substatMask)
	if num < 0 || num >= len(substatDefs) {
		return 0
	}
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	return template.SubstatWeight[substatDefs[num].NameCN]
}

func maxEchoScore(resonator, cost string) float64 {
	if cost == "" {
		cost = "1C"
	}
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	if len(cost) == 0 {
		return 0
	}
	return template.EchoMaxScore[cost[:1]]
}

func scorePercentile(samples []simulatorSample, echo EchoLog, resonator, cost string) float64 {
	stage := filledSubstatSlots(echo)
	current := scoreEcho(echo, resonator, cost).SubstatAll
	if current <= 0 {
		return 0
	}
	total := 0
	lowerOrEqual := 0
	for _, sample := range samples {
		item := EchoLog{
			Substat1: sample.Substats[0],
			Substat2: sample.Substats[1],
			Substat3: sample.Substats[2],
			Substat4: sample.Substats[3],
			Substat5: sample.Substats[4],
		}
		if filledSubstatSlots(item) != stage {
			continue
		}
		total++
		if scoreEcho(item, resonator, cost).SubstatAll <= current {
			lowerOrEqual++
		}
	}
	if total == 0 {
		return 0
	}
	return rounded(float64(lowerOrEqual)/float64(total), 4)
}

func defaultTargetBits(resonator, goal string) int64 {
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	type weightedSubstat struct {
		num    int
		weight float64
	}
	weights := make([]weightedSubstat, 0, len(substatDefs))
	for _, def := range substatDefs {
		weights = append(weights, weightedSubstat{
			num:    def.Number,
			weight: template.SubstatWeight[def.NameCN],
		})
	}
	slices.SortFunc(weights, func(a, b weightedSubstat) int {
		if a.weight == b.weight {
			return a.num - b.num
		}
		if a.weight > b.weight {
			return -1
		}
		return 1
	})
	count := 3
	switch strings.TrimSpace(goal) {
	case "保底":
		count = 2
	case "大毕业", "毕业":
		count = 4
	case "神品":
		count = 5
	}
	var bits int64
	for i := 0; i < len(weights) && i < count; i++ {
		if weights[i].weight <= 0 {
			continue
		}
		bits |= int64(1 << weights[i].num)
	}
	if bits == 0 {
		return 0b11
	}
	return bits
}

func goalThreshold(goal string, maxScore float64) float64 {
	switch strings.TrimSpace(goal) {
	case "保底":
		return maxScore * 0.72
	case "小毕业":
		return maxScore * 0.82
	case "神品":
		return maxScore * 0.95
	case "大毕业", "毕业":
		return maxScore * 0.9
	default:
		return maxScore * 0.9
	}
}

func isGoalMet(e EchoLog, score float64, targetBits int64, threshold float64) bool {
	if targetBits > 0 && (e.SubstatAll&targetBits) != targetBits {
		return false
	}
	return score >= threshold
}

func scoreBucket(score, maxScore float64) string {
	if maxScore <= 0 {
		return "止步中等"
	}
	ratio := score / maxScore
	switch {
	case ratio >= 0.95:
		return "神品"
	case ratio >= 0.9:
		return "大毕业"
	case ratio >= 0.82:
		return "小毕业"
	default:
		return "止步中等"
	}
}

func decisionRecommendation(nextProb, finishProb, lockedValue, percentile float64) string {
	switch {
	case finishProb >= 0.4 || (finishProb >= 0.25 && lockedValue >= 0.82 && percentile >= 0.75):
		return "continue_to_end"
	case nextProb >= 0.3 && finishProb >= 0.12:
		return "continue_once"
	case nextProb < 0.2 && lockedValue < 0.7:
		return "stop"
	default:
		return "high_risk"
	}
}

func buildDecisionReasons(e EchoLog, targetBits int64, nextProb, finishProb, percentile float64) []string {
	reasons := make([]string, 0, 3)
	if (e.SubstatAll & 0b11) == 0b11 {
		reasons = append(reasons, "当前已具备双暴，基础盘不差")
	} else {
		reasons = append(reasons, "当前还未形成双暴，后续命中压力更大")
	}
	if targetBits > 0 {
		reasons = append(reasons, "下一手命中目标词条的概率约为 "+formatRate(nextProb))
	}
	if percentile >= 0.8 {
		reasons = append(reasons, "同阶段历史分位较高，属于值得继续观察的样本")
	} else if finishProb >= 0.2 {
		reasons = append(reasons, "继续到底仍有一定成型概率，但风险已经开始放大")
	} else {
		reasons = append(reasons, "继续到底的达标率偏低，止损通常更稳")
	}
	return reasons
}

func formatRate(v float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v*100, 'f', 2, 64), "0"), ".") + "%"
}

func boolRate(ok bool) float64 {
	if ok {
		return 1
	}
	return 0
}
