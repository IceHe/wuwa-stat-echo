package goapp

import (
	"slices"
	"strings"
)

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
