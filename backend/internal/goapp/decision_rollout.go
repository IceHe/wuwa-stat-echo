package goapp

import (
	"math"
	"math/rand"
	"slices"
)

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
