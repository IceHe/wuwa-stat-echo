package goapp

import (
	"encoding/json"
	"math"
)

func bitCount(bits int64) int {
	count := 0
	for bits != 0 {
		count++
		bits &= bits - 1
	}
	return count
}

func bitPos(bits int64) int {
	pos := 0
	for bits != 0 {
		if bits&1 == 1 {
			return pos
		}
		bits >>= 1
		pos++
	}
	return -1
}

func rounded(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Round(val*p) / p
}

func cloneTuneStats(src *TuneStatsResponse) *TuneStatsResponse {
	if src == nil {
		return nil
	}
	buf, _ := json.Marshal(src)
	var out TuneStatsResponse
	_ = json.Unmarshal(buf, &out)
	return &out
}

func currentPos(e EchoLog) int {
	pos := 0
	if e.Substat1 != 0 {
		pos = 1
	}
	if e.Substat2 != 0 {
		pos = 2
	}
	if e.Substat3 != 0 {
		pos = 3
	}
	if e.Substat4 != 0 {
		pos = 4
	}
	return pos
}

func firstTierForMask(substats []int64, mask int64) int64 {
	for _, substat := range substats {
		if substat&mask != 0 {
			return substat >> substatBitWidth
		}
	}
	return 0
}

func incrementPredictCount(counts []int, bits int64) {
	index := bitPos(bits)
	if index >= 0 && index < len(counts) {
		counts[index]++
	}
}

func computeEchoLogsAnalysisFromItems(items []EchoLog, total int64, targetBits int64) map[string]any {
	found := false
	idx := 0
	targetCount := 0
	targetEchoDistance := -1
	targetSubstatDistance := -1
	substatTotal := 0
	tunerRecycled := 0
	expTotal := 0
	expRecycled := 0
	for _, echoLog := range items {
		substatAll := (echoLog.Substat1 | echoLog.Substat2 | echoLog.Substat3 | echoLog.Substat4 | echoLog.Substat5) & substatMask
		substatCount := bitCount(substatAll)
		substatTotal += substatCount
		expTotal += expTable[0][substatCount]
		if substatAll&targetBits == targetBits {
			targetCount++
			if !found {
				found = true
				targetEchoDistance = idx
				targetSubstatDistance = substatTotal
			}
		} else {
			tunerRecycled += substatCount * tunerRecycledPerSubstat
			expRecycled += expReturn[substatCount]
		}
		idx++
	}
	if !found {
		targetEchoDistance = idx
		targetSubstatDistance = substatTotal
	}
	tunerConsumed := int(math.Ceil(float64(substatTotal*10 - tunerRecycled)))
	expConsumed := int(math.Ceil(float64(expTotal-expRecycled) / expGold))
	resp := map[string]any{
		"sample_size":             total,
		"target_echo_distance":    targetEchoDistance,
		"target_substat_distance": targetSubstatDistance,
		"target":                  targetCount,
		"target_avg_echo":         0.0,
		"target_avg_substat":      0.0,
		"tuner_consumed":          tunerConsumed,
		"tuner_consumed_avg":      0.0,
		"exp_consumed":            expConsumed,
		"exp_consumed_avg":        0.0,
		"target_rate_stats":       newProportionStat(int64(targetCount), total),
	}
	if targetCount > 0 {
		resp["target_avg_echo"] = rounded(float64(total)/float64(targetCount), 1)
		resp["target_avg_substat"] = rounded(float64(substatTotal)/float64(targetCount), 1)
		resp["tuner_consumed_avg"] = int(math.Ceil(float64(tunerConsumed) / float64(targetCount)))
		resp["exp_consumed_avg"] = int(math.Ceil(float64(expConsumed) / float64(targetCount)))
	}
	return resp
}
