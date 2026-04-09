package goapp

import (
	"strconv"
	"strings"
)

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
