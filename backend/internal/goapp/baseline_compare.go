package goapp

import (
	"math"
	"sort"
)

type SignificanceSummary struct {
	Significant  bool    `json:"significant"`
	SampleEnough bool    `json:"sample_enough"`
	PValue       float64 `json:"p_value"`
	ZScore       float64 `json:"z_score"`
	EffectSizePP float64 `json:"effect_size_pp"`
	Direction    string  `json:"direction"`
}

type BiasHintSummary struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RateComparison struct {
	User         *ProportionStat      `json:"user"`
	Global       *ProportionStat      `json:"global"`
	DeltaRate    float64              `json:"delta_rate"`
	DeltaCount   int64                `json:"delta_count"`
	Significance *SignificanceSummary `json:"significance,omitempty"`
	BiasHint     *BiasHintSummary     `json:"bias_hint,omitempty"`
}

func buildRateComparison(userStat *ProportionStat, globalStat *ProportionStat) *RateComparison {
	if userStat == nil || globalStat == nil {
		return nil
	}
	deltaRate := rounded(userStat.Rate-globalStat.Rate, 2)
	return &RateComparison{
		User:         userStat,
		Global:       globalStat,
		DeltaRate:    deltaRate,
		DeltaCount:   userStat.Count - globalStat.Count,
		Significance: buildSignificanceSummary(userStat, globalStat),
		BiasHint:     buildBiasHint(userStat, globalStat, deltaRate),
	}
}

func buildTuneStatsBaselineCompare(userStats *TuneStatsResponse, globalStats *TuneStatsResponse) map[string]any {
	if userStats == nil || globalStats == nil {
		return nil
	}
	substatRateDelta := map[string]*RateComparison{}
	type highlight struct {
		Key   string
		Name  string
		Delta float64
		Comp  *RateComparison
	}
	highlights := make([]highlight, 0)
	for key, userItem := range userStats.SubstatDict {
		globalItem := globalStats.SubstatDict[key]
		if userItem == nil || globalItem == nil {
			continue
		}
		comp := buildRateComparison(userItem.Proportion, globalItem.Proportion)
		substatRateDelta[key] = comp
		if comp != nil && comp.BiasHint != nil && comp.BiasHint.Code != "no_clear_difference" {
			highlights = append(highlights, highlight{
				Key:   key,
				Name:  userItem.NameCN,
				Delta: math.Abs(comp.DeltaRate),
				Comp:  comp,
			})
		}
	}
	sort.Slice(highlights, func(i, j int) bool { return highlights[i].Delta > highlights[j].Delta })
	if len(highlights) > 5 {
		highlights = highlights[:5]
	}
	highlightItems := make([]map[string]any, 0, len(highlights))
	for _, item := range highlights {
		highlightItems = append(highlightItems, map[string]any{
			"substat":    item.Key,
			"name_cn":    item.Name,
			"comparison": item.Comp,
		})
	}
	return map[string]any{
		"user_sample_size":   userStats.DataTotal,
		"global_sample_size": globalStats.DataTotal,
		"substat_rate_delta": substatRateDelta,
		"highlights":         highlightItems,
	}
}

func buildSignificanceSummary(userStat *ProportionStat, globalStat *ProportionStat) *SignificanceSummary {
	if userStat == nil || globalStat == nil || userStat.Total <= 0 || globalStat.Total <= 0 {
		return nil
	}
	p1 := float64(userStat.Count) / float64(userStat.Total)
	p2 := float64(globalStat.Count) / float64(globalStat.Total)
	pooled := float64(userStat.Count+globalStat.Count) / float64(userStat.Total+globalStat.Total)
	denom := math.Sqrt(pooled * (1 - pooled) * ((1 / float64(userStat.Total)) + (1 / float64(globalStat.Total))))
	zScore := 0.0
	pValue := 1.0
	if denom > 0 {
		zScore = (p1 - p2) / denom
		pValue = math.Erfc(math.Abs(zScore) / math.Sqrt2)
	}
	direction := "similar"
	if p1 > p2 {
		direction = "higher"
	} else if p1 < p2 {
		direction = "lower"
	}
	return &SignificanceSummary{
		Significant:  pValue < 0.05,
		SampleEnough: userStat.Total >= 30 && globalStat.Total >= 30,
		PValue:       rounded(pValue, 4),
		ZScore:       rounded(zScore, 3),
		EffectSizePP: rounded((p1-p2)*100, 2),
		Direction:    direction,
	}
}

func buildBiasHint(userStat *ProportionStat, globalStat *ProportionStat, deltaRate float64) *BiasHintSummary {
	significance := buildSignificanceSummary(userStat, globalStat)
	if significance == nil {
		return nil
	}
	if !significance.SampleEnough {
		return &BiasHintSummary{Code: "sample_too_small", Message: "样本量偏小，先观察趋势，不建议下结论"}
	}
	if significance.Significant {
		if deltaRate >= 2 {
			return &BiasHintSummary{Code: "significantly_higher", Message: "个人结果显著高于全站基线"}
		}
		if deltaRate <= -2 {
			return &BiasHintSummary{Code: "significantly_lower", Message: "个人结果显著低于全站基线"}
		}
		if deltaRate > 0 {
			return &BiasHintSummary{Code: "slightly_higher", Message: "个人结果高于全站，且差异达到统计显著"}
		}
		if deltaRate < 0 {
			return &BiasHintSummary{Code: "slightly_lower", Message: "个人结果低于全站，且差异达到统计显著"}
		}
	}
	if math.Abs(deltaRate) >= 2 {
		return &BiasHintSummary{Code: "difference_not_significant", Message: "和全站存在差距，但暂未达到统计显著"}
	}
	return &BiasHintSummary{Code: "no_clear_difference", Message: "和全站相比暂无明确偏差"}
}
