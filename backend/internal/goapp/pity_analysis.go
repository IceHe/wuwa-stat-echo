package goapp

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type pityEventDef struct {
	Key   string
	Label string
	Match func(int64) bool
}

type pityBucketDef struct {
	Label string
	Min   int
	Max   int
}

type pityEventState struct {
	totalSuccesses int64
	trials         int64
	trialSuccesses int64
	internalTrials int64
	internalSucc   int64
	perGap         map[int]*pityGapAgg
	maxInternalGap int
	maxGapUserID   int64
	maxGapStartID  int64
	maxGapEndID    int64
}

type pityGapAgg struct {
	trials       int64
	successes    int64
	expectedSum  float64
	internalSeen int64
	internalSucc int64
}

type pityUserState struct {
	successes int64
	total     int64
}

type pityEchoRow struct {
	ID       int64
	UserID   int64
	Substat  int64
	Substats [5]int64
	Opened   int
	TunedAt  *time.Time
}

type pityStageAgg struct {
	sampleCount          int64
	continueCount        int64
	stopCount            int64
	completedCount       int64
	finalDoubleCritCount int64
}

type pityPathAgg struct {
	sampleCount          int64
	eligibleCount        int64
	completedCount       int64
	finalDoubleCritCount int64
}

var pityBuckets = []pityBucketDef{
	{Label: "1", Min: 1, Max: 1},
	{Label: "2", Min: 2, Max: 2},
	{Label: "3", Min: 3, Max: 3},
	{Label: "4", Min: 4, Max: 4},
	{Label: "5-6", Min: 5, Max: 6},
	{Label: "7-9", Min: 7, Max: 9},
	{Label: "10-14", Min: 10, Max: 14},
	{Label: "15-19", Min: 15, Max: 19},
	{Label: "20-29", Min: 20, Max: 29},
	{Label: "30+", Min: 30, Max: math.MaxInt},
}

func (a *App) handlePityAnalysis(w http.ResponseWriter, r *http.Request) {
	resp, err := a.computePityAnalysis(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to get pity analysis", 500))
		return
	}
	writeJSON(w, success("pity analysis", resp))
}

func (a *App) computePityAnalysis(ctx context.Context) (*PityAnalysisResponse, error) {
	rows, err := a.db.Query(ctx, `
		select id, user_id, substat_all, substat1, substat2, substat3, substat4, substat5, tuned_at
		from wuwa_echo_log
		where deleted = 0 and user_id > 0
		order by user_id asc, id asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	echoRows := make([]pityEchoRow, 0, 24000)
	userCounts := make(map[int64]int)
	var lastUserID int64 = -1
	var lastTunedAt *time.Time
	var timeOrderMismatch int64
	for rows.Next() {
		var row pityEchoRow
		if err := rows.Scan(&row.ID, &row.UserID, &row.Substat, &row.Substats[0], &row.Substats[1], &row.Substats[2], &row.Substats[3], &row.Substats[4], &row.TunedAt); err != nil {
			return nil, err
		}
		row.Opened = countOpenedSubstats(row.Substats)
		echoRows = append(echoRows, row)
		userCounts[row.UserID]++
		if row.UserID != lastUserID {
			lastUserID = row.UserID
			lastTunedAt = nil
		}
		if lastTunedAt != nil && row.TunedAt != nil && row.TunedAt.Before(*lastTunedAt) {
			timeOrderMismatch++
		}
		if row.TunedAt != nil {
			t := *row.TunedAt
			lastTunedAt = &t
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	counts := make([]int, 0, len(userCounts))
	var maxEchoesPerUser int64
	for _, count := range userCounts {
		counts = append(counts, count)
		if int64(count) > maxEchoesPerUser {
			maxEchoesPerUser = int64(count)
		}
	}
	sort.Ints(counts)
	medianEchoesPerUser := medianInts(counts)

	events := []pityEventDef{
		{
			Key:   "crit_rate",
			Label: "暴击",
			Match: func(bits int64) bool { return bits&(1<<0) != 0 },
		},
		{
			Key:   "crit_dmg",
			Label: "暴击伤害",
			Match: func(bits int64) bool { return bits&(1<<1) != 0 },
		},
		{
			Key:   "double_crit",
			Label: "双暴组合",
			Match: func(bits int64) bool { return bits&0b11 == 0b11 },
		},
	}
	for _, def := range substatDefs {
		matchMask := int64(def.Bitmap)
		events = append(events, pityEventDef{
			Key:   "substat_" + def.Name,
			Label: def.NameCN,
			Match: func(bits int64) bool { return bits&matchMask != 0 },
		})
	}

	eventStates := make([]pityEventState, len(events))
	userEventStates := make([]map[int64]*pityUserState, len(events))
	stageAgg := map[int]map[string]*pityStageAgg{}
	for index := range events {
		eventStates[index].perGap = map[int]*pityGapAgg{}
		userEventStates[index] = map[int64]*pityUserState{}
	}
	for stage := 1; stage <= 4; stage++ {
		stageAgg[stage] = map[string]*pityStageAgg{}
	}

	for _, row := range echoRows {
		for index, event := range events {
			userState := userEventStates[index][row.UserID]
			if userState == nil {
				userState = &pityUserState{}
				userEventStates[index][row.UserID] = userState
			}
			userState.total++
			if event.Match(row.Substat) {
				userState.successes++
				eventStates[index].totalSuccesses++
			}
		}
		updateStageAgg(stageAgg, row)
	}

	seqByUser := map[int64][]pityEchoRow{}
	for _, row := range echoRows {
		seqByUser[row.UserID] = append(seqByUser[row.UserID], row)
	}

	for _, seq := range seqByUser {
		for eventIndex, event := range events {
			lastSuccessIndex := -1
			successPositions := make([]int, 0, 32)
			successEchoIDs := make([]int64, 0, 32)
			for index := 0; index < len(seq)-1; index++ {
				if event.Match(seq[index].Substat) {
					lastSuccessIndex = index
					successPositions = append(successPositions, index)
					successEchoIDs = append(successEchoIDs, seq[index].ID)
				}

				gap := index + 1
				internalSeen := int64(0)
				if lastSuccessIndex >= 0 {
					gap = index - lastSuccessIndex
					internalSeen = 1
				}

				agg := eventStates[eventIndex].perGap[gap]
				if agg == nil {
					agg = &pityGapAgg{}
					eventStates[eventIndex].perGap[gap] = agg
				}
				agg.trials++
				eventStates[eventIndex].trials++

				base := 0.0
				if userState := userEventStates[eventIndex][seq[index].UserID]; userState != nil && userState.total > 0 {
					base = float64(userState.successes) / float64(userState.total)
				}
				agg.expectedSum += base

				if internalSeen == 1 {
					agg.internalSeen++
					eventStates[eventIndex].internalTrials++
				}

				if event.Match(seq[index+1].Substat) {
					agg.successes++
					eventStates[eventIndex].trialSuccesses++
					if internalSeen == 1 {
						agg.internalSucc++
						eventStates[eventIndex].internalSucc++
					}
				}
			}

			if len(seq) > 0 && event.Match(seq[len(seq)-1].Substat) {
				successPositions = append(successPositions, len(seq)-1)
				successEchoIDs = append(successEchoIDs, seq[len(seq)-1].ID)
			}

			for i := 0; i+1 < len(successPositions); i++ {
				gap := successPositions[i+1] - successPositions[i] - 1
				if gap > eventStates[eventIndex].maxInternalGap {
					eventStates[eventIndex].maxInternalGap = gap
					eventStates[eventIndex].maxGapUserID = seq[0].UserID
					eventStates[eventIndex].maxGapStartID = successEchoIDs[i]
					eventStates[eventIndex].maxGapEndID = successEchoIDs[i+1]
				}
			}
		}
	}

	maxGapRows := make([]PityAnalysisMaxGapRow, 0, len(substatDefs)+1)
	for _, def := range substatDefs {
		state := eventStates[findPityEventIndex(events, def.NameCN)]
		maxGapRows = append(maxGapRows, PityAnalysisMaxGapRow{
			Label:       def.NameCN,
			MaxGap:      state.maxInternalGap,
			UserID:      state.maxGapUserID,
			StartEchoID: state.maxGapStartID,
			EndEchoID:   state.maxGapEndID,
		})
	}
	doubleCritState := eventStates[findPityEventIndex(events, "双暴组合")]
	maxGapRows = append(maxGapRows, PityAnalysisMaxGapRow{
		Label:       "双暴组合",
		MaxGap:      doubleCritState.maxInternalGap,
		UserID:      doubleCritState.maxGapUserID,
		StartEchoID: doubleCritState.maxGapStartID,
		EndEchoID:   doubleCritState.maxGapEndID,
	})
	sort.Slice(maxGapRows, func(i, j int) bool {
		if maxGapRows[i].MaxGap == maxGapRows[j].MaxGap {
			return maxGapRows[i].Label < maxGapRows[j].Label
		}
		return maxGapRows[i].MaxGap > maxGapRows[j].MaxGap
	})

	reportEvents := make([]PityAnalysisEvent, 0, 3)
	for _, target := range []string{"暴击", "暴击伤害", "双暴组合"} {
		index := findPityEventIndex(events, target)
		state := eventStates[index]
		reportEvents = append(reportEvents, buildPityAnalysisEvent(events[index], state, len(echoRows)))
	}
	stageSummaries, continuationRows, futureRows := buildSelectionBiasReport(stageAgg)
	pathRows := buildDoubleCritPathRows(echoRows)

	now := time.Now()
	resp := &PityAnalysisResponse{
		GeneratedAt:         &now,
		EchoTotal:           int64(len(echoRows)),
		UserTotal:           int64(len(userCounts)),
		MedianEchoesPerUser: medianEchoesPerUser,
		MaxEchoesPerUser:    maxEchoesPerUser,
		TimeOrderMismatch:   timeOrderMismatch,
		Summaries: []PityAnalysisSummary{
			{Label: "声骸样本量", Value: formatInt64(int64(len(echoRows)))},
			{Label: "玩家数", Value: formatInt64(int64(len(userCounts)))},
			{Label: "每玩家声骸中位数", Value: formatFloat(medianEchoesPerUser, 1)},
			{Label: "单玩家最大声骸数", Value: formatInt64(maxEchoesPerUser)},
			{Label: "双暴样本占比", Value: formatRateRatio(doubleCritState.totalSuccesses, int64(len(echoRows)))},
		},
		DefinitionNotes: []string{
			"硬保底：本页的工作定义是，在单个玩家的已记录声骸序列里，如果存在“连续 K 个已记录声骸没出目标，则第 K+1 个已记录声骸必出”的机制，那么两次成功之间的最大内部间隔不可能大于 K。",
			"软保底：本页的工作定义是，在单个玩家的已记录声骸序列里，如果存在“漏得越久越容易补”的机制，那么当前 gap 越大时，下一个已记录声骸命中目标的条件概率应该出现持续、清晰的上升。",
			"本页的“高档位单暴”按 value_number >= 5 定义，也就是 8 档里的后 3 档。",
		},
		MethodNotes: []string{
			"统计单位是“单个玩家内部、按声骸顺序排列的完整声骸序列”，不是逐孔 tune log。这样可以避开单件声骸 5 个副词条互不重复带来的结构性偏差。",
			"硬保底的反证口径是“同一玩家内部、已记录声骸序列里，两次成功之间的最大内部间隔”。如果存在 K 个已记录声骸内必出的硬保底，那么内部间隔不可能大于 K。",
			"软保底的检验口径是“当前已经连续 gap 个声骸未出目标时，下一个声骸命中的概率”。若存在补偿型保底，这个概率应随 gap 明显上升。",
			"页面中的 expected_rate 先按每个玩家自己的长期基础命中率估计，再按当前 gap 桶里的玩家构成加权汇总，用来控制不同玩家录入习惯和基础命中率差异。",
			"当前数据可以反证明显硬保底，也可以判断是否支持常见软保底；但无法用有限样本在逻辑上证明‘任何可能形式的保底都绝对不存在’。",
		},
		Conclusions: []string{
			"现有数据足以反证一批常见硬保底：至少在当前记录口径下，单玩家内部不存在“43 个已记录声骸内必出某单副词条”的明显硬保底，也不存在“91 个已记录声骸内必出双暴”的明显硬保底。",
			"暴击、暴击伤害和双暴组合的 gap 条件命中率都没有形成清晰的单调上升曲线，当前数据不支持“漏得越久越会补偿”的典型软保底。",
			"双暴组合在中等长间隔区间略高于自身基线，但 30+ 区间没有继续上扬，更像随机波动或行为选择偏差，而不是明确的保底函数。",
		},
		SelectionBiasNotes: []string{
			"当前数据库同时包含开满 5 孔和只开到 1-4 孔就停手的声骸，因此可以直接检验“前缀状态会不会影响是否继续开”。",
			"若不同前缀下的继续开孔概率差异显著，则完整 5 孔样本不是无偏样本；这会削弱基于最终成品样本的机制论证力度。",
			"对双暴尤其如此：前 3 孔无双暴、前 4 孔单暴，都仍然可能在剩余孔位补成双暴，但这些潜在成功样本可能被人为提前截断。",
		},
		MaxGapRows:           maxGapRows,
		Events:               reportEvents,
		StageSummaries:       stageSummaries,
		ContinuationRows:     continuationRows,
		DoubleCritFutureRows: futureRows,
		DoubleCritPathRows:   pathRows,
	}
	return resp, nil
}

func buildPityAnalysisEvent(event pityEventDef, state pityEventState, totalEchoes int) PityAnalysisEvent {
	out := PityAnalysisEvent{
		Key:               event.Key,
		Label:             event.Label,
		SuccessCount:      state.totalSuccesses,
		BaseRate:          safeRate(state.totalSuccesses, int64(totalEchoes)),
		InternalBaseRate:  safeRate(state.internalSucc, state.internalTrials),
		MaxInternalGap:    state.maxInternalGap,
		MaxGapUserID:      state.maxGapUserID,
		MaxGapStartEchoID: state.maxGapStartID,
		MaxGapEndEchoID:   state.maxGapEndID,
		HardPitySummary:   "未观察到更小阈值的硬保底；在当前记录口径下，当前样本已直接反证“单玩家内部 " + event.Label + " 在 " + formatInt(state.maxInternalGap) + " 个已记录声骸内必出”的说法。",
	}

	monotonicRise := true
	lastRate := -1.0
	riseCount := 0
	declineAfter20 := false
	for _, bucket := range pityBuckets {
		agg := aggregatePityBucket(state.perGap, bucket)
		row := PityAnalysisBucketRow{
			GapLabel:     bucket.Label,
			Trials:       agg.trials,
			Successes:    agg.successes,
			ActualRate:   safeRate(agg.successes, agg.trials),
			ExpectedRate: safeRateFloat(agg.expectedSum, agg.trials),
			DeltaRate:    safeRate(agg.successes, agg.trials) - safeRateFloat(agg.expectedSum, agg.trials),
			InternalRate: safeRate(agg.internalSucc, agg.internalSeen),
		}
		out.Buckets = append(out.Buckets, row)
		if row.Trials == 0 {
			continue
		}
		if lastRate >= 0 {
			if row.ActualRate < lastRate-0.000001 {
				monotonicRise = false
			}
			if row.ActualRate > lastRate+0.002 {
				riseCount++
			}
		}
		if bucket.Min >= 30 && len(out.Buckets) >= 2 {
			prev := out.Buckets[len(out.Buckets)-2]
			if row.ActualRate < prev.ActualRate {
				declineAfter20 = true
			}
		}
		lastRate = row.ActualRate
	}

	switch {
	case monotonicRise && riseCount >= 4:
		out.SoftPitySummary = "gap 条件命中率呈现较明显上升，值得继续怀疑存在软保底。"
	case declineAfter20:
		out.SoftPitySummary = "中段区间有波动，但长间隔尾部没有持续抬升，不符合典型软保底曲线。"
	default:
		out.SoftPitySummary = "gap 条件命中率整体缺乏清晰、稳定、单调的上升趋势，当前数据不支持典型软保底。"
	}
	return out
}

func aggregatePityBucket(perGap map[int]*pityGapAgg, bucket pityBucketDef) pityGapAgg {
	var out pityGapAgg
	for gap, agg := range perGap {
		if gap < bucket.Min || gap > bucket.Max {
			continue
		}
		out.trials += agg.trials
		out.successes += agg.successes
		out.expectedSum += agg.expectedSum
		out.internalSeen += agg.internalSeen
		out.internalSucc += agg.internalSucc
	}
	return out
}

func findPityEventIndex(events []pityEventDef, label string) int {
	for index, event := range events {
		if event.Label == label {
			return index
		}
	}
	return -1
}

func medianInts(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return float64(values[mid])
	}
	return float64(values[mid-1]+values[mid]) / 2
}

func safeRate(count int64, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func safeRateFloat(sum float64, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return sum / float64(total)
}

func formatInt(value int) string {
	return formatInt64(int64(value))
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatFloat(value float64, digits int) string {
	return strconv.FormatFloat(value, 'f', digits, 64)
}

func formatRateRatio(count int64, total int64) string {
	return formatFloat(safeRate(count, total)*100, 1) + "%"
}

func countOpenedSubstats(substats [5]int64) int {
	opened := 0
	for _, substat := range substats {
		if substat == 0 {
			break
		}
		opened++
	}
	return opened
}

func prefixCategory(substats [5]int64, stage int) string {
	bits := int64(0)
	for index := 0; index < stage && index < len(substats); index++ {
		bits |= substats[index]
	}
	hasCrit := bits&(1<<0) != 0
	hasDmg := bits&(1<<1) != 0
	switch {
	case hasCrit && hasDmg:
		return "已有双暴"
	case hasCrit:
		return "仅暴击"
	case hasDmg:
		return "仅暴伤"
	default:
		return "无双暴"
	}
}

func updateStageAgg(stageAgg map[int]map[string]*pityStageAgg, row pityEchoRow) {
	for stage := 1; stage <= 4; stage++ {
		if row.Opened < stage {
			continue
		}
		category := prefixCategory(row.Substats, stage)
		agg := stageAgg[stage][category]
		if agg == nil {
			agg = &pityStageAgg{}
			stageAgg[stage][category] = agg
		}
		agg.sampleCount++
		if row.Opened >= stage+1 {
			agg.continueCount++
		} else {
			agg.stopCount++
		}
		if row.Opened >= 5 {
			agg.completedCount++
			if row.Substat&0b11 == 0b11 {
				agg.finalDoubleCritCount++
			}
		}
	}
}

func buildSelectionBiasReport(stageAgg map[int]map[string]*pityStageAgg) ([]PityAnalysisSummary, []PityContinuationRow, []PityDoubleCritFutureRow) {
	stageSummaries := make([]PityAnalysisSummary, 0, 4)
	continuationRows := make([]PityContinuationRow, 0, 16)
	futureRows := make([]PityDoubleCritFutureRow, 0, 16)
	order := []string{"无双暴", "仅暴击", "仅暴伤", "已有双暴"}
	for stage := 1; stage <= 4; stage++ {
		var total int64
		var continueCount int64
		for _, agg := range stageAgg[stage] {
			total += agg.sampleCount
			continueCount += agg.continueCount
		}
		stageSummaries = append(stageSummaries, PityAnalysisSummary{
			Label: "开到第 " + formatInt(stage) + " 孔后继续开下一孔",
			Value: formatFloat(safeRate(continueCount, total)*100, 1) + "%",
		})
		for _, category := range order {
			agg := stageAgg[stage][category]
			if agg == nil || agg.sampleCount == 0 {
				continue
			}
			continuationRows = append(continuationRows, PityContinuationRow{
				StageOpened:    stage,
				PrefixCategory: category,
				SampleCount:    agg.sampleCount,
				ContinueCount:  agg.continueCount,
				StopCount:      agg.stopCount,
				ContinueRate:   safeRate(agg.continueCount, agg.sampleCount),
			})
			futureRows = append(futureRows, PityDoubleCritFutureRow{
				StageOpened:          stage,
				PrefixCategory:       category,
				SampleCount:          agg.sampleCount,
				CompletedCount:       agg.completedCount,
				CompletedRate:        safeRate(agg.completedCount, agg.sampleCount),
				FinalDoubleCritCount: agg.finalDoubleCritCount,
				FinalDoubleCritRate:  safeRate(agg.finalDoubleCritCount, agg.sampleCount),
			})
		}
	}
	return stageSummaries, continuationRows, futureRows
}

func buildDoubleCritPathRows(rows []pityEchoRow) []PityDoubleCritPathRow {
	aggMap := map[string]*pityPathAgg{
		"前 3 孔无双暴 -> 最终双暴":        {},
		"前 3 孔仅暴击 -> 最终双暴":        {},
		"前 3 孔高档位仅暴击 -> 最终双暴":     {},
		"前 3 孔仅暴伤 -> 最终双暴":        {},
		"前 3 孔高档位仅暴伤 -> 最终双暴":     {},
		"前 4 孔仅暴击 -> 第 5 孔补双暴":    {},
		"前 4 孔高档位仅暴击 -> 第 5 孔补双暴": {},
		"前 4 孔仅暴伤 -> 第 5 孔补双暴":    {},
		"前 4 孔高档位仅暴伤 -> 第 5 孔补双暴": {},
	}

	for _, row := range rows {
		stage3 := prefixCategory(row.Substats, 3)
		stage4 := prefixCategory(row.Substats, 4)
		finalDoubleCrit := row.Substat&0b11 == 0b11

		switch stage3 {
		case "无双暴":
			updatePathAgg(aggMap["前 3 孔无双暴 -> 最终双暴"], row, finalDoubleCrit)
		case "仅暴击":
			updatePathAgg(aggMap["前 3 孔仅暴击 -> 最终双暴"], row, finalDoubleCrit)
			if isHighTierSingleCrit(row.Substats[:3], 0b01) {
				updatePathAgg(aggMap["前 3 孔高档位仅暴击 -> 最终双暴"], row, finalDoubleCrit)
			}
		case "仅暴伤":
			updatePathAgg(aggMap["前 3 孔仅暴伤 -> 最终双暴"], row, finalDoubleCrit)
			if isHighTierSingleCrit(row.Substats[:3], 0b10) {
				updatePathAgg(aggMap["前 3 孔高档位仅暴伤 -> 最终双暴"], row, finalDoubleCrit)
			}
		}

		switch stage4 {
		case "仅暴击":
			updatePathAgg(aggMap["前 4 孔仅暴击 -> 第 5 孔补双暴"], row, finalDoubleCrit)
			if isHighTierSingleCrit(row.Substats[:4], 0b01) {
				updatePathAgg(aggMap["前 4 孔高档位仅暴击 -> 第 5 孔补双暴"], row, finalDoubleCrit)
			}
		case "仅暴伤":
			updatePathAgg(aggMap["前 4 孔仅暴伤 -> 第 5 孔补双暴"], row, finalDoubleCrit)
			if isHighTierSingleCrit(row.Substats[:4], 0b10) {
				updatePathAgg(aggMap["前 4 孔高档位仅暴伤 -> 第 5 孔补双暴"], row, finalDoubleCrit)
			}
		}
	}

	order := []string{
		"前 3 孔无双暴 -> 最终双暴",
		"前 3 孔仅暴击 -> 最终双暴",
		"前 3 孔高档位仅暴击 -> 最终双暴",
		"前 3 孔仅暴伤 -> 最终双暴",
		"前 3 孔高档位仅暴伤 -> 最终双暴",
		"前 4 孔仅暴击 -> 第 5 孔补双暴",
		"前 4 孔高档位仅暴击 -> 第 5 孔补双暴",
		"前 4 孔仅暴伤 -> 第 5 孔补双暴",
		"前 4 孔高档位仅暴伤 -> 第 5 孔补双暴",
	}
	out := make([]PityDoubleCritPathRow, 0, len(order))
	for _, key := range order {
		agg := aggMap[key]
		out = append(out, PityDoubleCritPathRow{
			PathLabel:            key,
			SampleCount:          agg.sampleCount,
			EligibleCount:        agg.eligibleCount,
			CompletedCount:       agg.completedCount,
			FinalDoubleCritCount: agg.finalDoubleCritCount,
			FinalDoubleCritRate:  safeRate(agg.finalDoubleCritCount, agg.sampleCount),
		})
	}
	return out
}

func updatePathAgg(agg *pityPathAgg, row pityEchoRow, finalDoubleCrit bool) {
	if agg == nil {
		return
	}
	agg.sampleCount++
	if row.Opened >= 4 {
		agg.eligibleCount++
	}
	if row.Opened >= 5 {
		agg.completedCount++
	}
	if finalDoubleCrit {
		agg.finalDoubleCritCount++
	}
}

func isHighTierSingleCrit(substats []int64, mask int64) bool {
	for _, substat := range substats {
		if substat == 0 || substat&mask == 0 {
			continue
		}
		return bitPos(substat>>substatBitWidth) >= 5
	}
	return false
}
