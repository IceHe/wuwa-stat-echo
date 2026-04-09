package goapp

import (
	"context"
	"fmt"
	"strconv"
)

func newSubstatDict() map[string]*SubstatItem {
	out := make(map[string]*SubstatItem, len(substatDefs))
	for _, def := range substatDefs {
		item := &SubstatItem{
			Number:        def.Number,
			Name:          def.Name,
			NameCN:        def.NameCN,
			ValueDict:     map[string]*SubstatValueStat{},
			CurPosPercent: "",
		}
		item.ValueDict["all"] = newSubstatValueStat(0, "all", "所有档位")
		for _, value := range def.Values {
			item.ValueDict[strconv.Itoa(value.ValueNumber)] = newSubstatValueStat(value.ValueNumber, value.ValueDesc, value.ValueFull)
		}
		out[strconv.Itoa(def.Number)] = item
	}
	return out
}

func newSubstatValueStat(valueNumber int, valueDesc, full string) *SubstatValueStat {
	return &SubstatValueStat{
		ValueNumber:   valueNumber,
		ValueDesc:     valueDesc,
		ValueDescFull: full,
		PositionDict: map[string]*SubstatValuePositionStat{
			"0": {Position: 0},
			"1": {Position: 1},
			"2": {Position: 2},
			"3": {Position: 3},
			"4": {Position: 4},
		},
	}
}

func (a *App) refreshCachedTuneStats(ctx context.Context) error {
	stats, err := a.loadTuneStatsFromAggregate(ctx, 0)
	if err != nil {
		return err
	}
	if stats == nil {
		stats, err = a.computeTuneStats(ctx, 0, 0, 0, 0, parseStatsWindow(""))
	}
	if err != nil {
		return err
	}
	a.statsMu.Lock()
	a.cachedStats = stats
	a.statsMu.Unlock()
	return nil
}

func (a *App) getCachedTuneStats() *TuneStatsResponse {
	a.statsMu.RLock()
	defer a.statsMu.RUnlock()
	return cloneTuneStats(a.cachedStats)
}

func (a *App) computeTuneStats(ctx context.Context, size int, userID int64, afterID int64, beforeID int64, window statsWindow) (*TuneStatsResponse, error) {
	substatDict := newSubstatDict()

	countSQL := "select count(id) from wuwa_tune_log where deleted = 0"
	var countArgs []any
	querySQL := "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where deleted = 0"
	var queryArgs []any
	argPos := 1

	addFilter := func(expr string, value any) {
		countSQL += fmt.Sprintf(" and %s $%d", expr, argPos)
		querySQL += fmt.Sprintf(" and %s $%d", expr, argPos)
		countArgs = append(countArgs, value)
		queryArgs = append(queryArgs, value)
		argPos++
	}
	if userID > 0 {
		addFilter("user_id =", userID)
	}
	if afterID > 0 {
		addFilter("id >", afterID)
	}
	if beforeID > 0 {
		addFilter("id <", beforeID)
	}
	if since := window.sinceTime(); since != nil {
		addFilter("timestamp >=", *since)
	}
	querySQL += " order by id desc"
	effectiveSize := window.applyLimit(size)
	if effectiveSize > 0 {
		querySQL += fmt.Sprintf(" limit %d", effectiveSize)
	}

	var logsTotal int64
	if err := a.db.QueryRow(ctx, countSQL, countArgs...).Scan(&logsTotal); err != nil {
		return nil, err
	}

	rows, err := a.db.Query(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []SubstatLog
	for rows.Next() {
		var logItem SubstatLog
		if err := rows.Scan(&logItem.ID, &logItem.Substat, &logItem.Value, &logItem.Position, &logItem.EchoID, &logItem.UserID, &logItem.OperatorID, &logItem.Timestamp, &logItem.Deleted); err != nil {
			return nil, err
		}
		logs = append(logs, logItem)
	}
	if effectiveSize > 0 {
		logsTotal = int64(len(logs))
	}

	distances := make([]int, 13)
	for i := range distances {
		distances[i] = -1
	}
	positionTotal := make([]int, 5)
	substatPosTotal := make([][]int, 13)
	for i := range substatPosTotal {
		substatPosTotal[i] = make([]int, 5)
	}

	index := -1
	for _, tuneLog := range logs {
		index++
		if tuneLog.Substat >= 0 && tuneLog.Substat < len(distances) && distances[tuneLog.Substat] == -1 {
			distances[tuneLog.Substat] = index
		}
		substat := substatDict[strconv.Itoa(tuneLog.Substat)]
		if substat == nil {
			continue
		}
		substat.Total++
		valueStat := substat.ValueDict[strconv.Itoa(tuneLog.Value)]
		if valueStat != nil {
			valueStat.Total++
			if posStat := valueStat.PositionDict[strconv.Itoa(tuneLog.Position)]; posStat != nil {
				posStat.Total++
			}
		}
		if tuneLog.Position >= 0 && tuneLog.Position < len(positionTotal) {
			positionTotal[tuneLog.Position]++
			substatPosTotal[tuneLog.Substat][tuneLog.Position]++
		}
		allStat := substat.ValueDict["all"]
		allStat.Total++
		if posStat := allStat.PositionDict[strconv.Itoa(tuneLog.Position)]; posStat != nil {
			posStat.Total++
		}
	}

	for _, substat := range substatDict {
		allStat := substat.ValueDict["all"]
		substat.Proportion = newProportionStat(int64(substat.Total), logsTotal)
		if logsTotal > 0 {
			substat.Percent = rounded(float64(substat.Total)/float64(logsTotal)*100, 2)
		}
		for key, value := range substat.ValueDict {
			denominator := int64(allStat.Total)
			if key == "all" {
				denominator = logsTotal
			}
			value.Proportion = newProportionStat(int64(value.Total), denominator)
			if substat.Total > 0 {
				value.PercentSubstat = rounded(float64(value.Total)/float64(substat.Total)*100, 2)
			}
			if key == "all" {
				if logsTotal > 0 {
					value.Percent = rounded(float64(value.Total)/float64(logsTotal)*100, 2)
				}
			} else if allStat.Total > 0 {
				value.Percent = rounded(float64(value.Total)/float64(allStat.Total)*100, 2)
			}
			for posKey, posStat := range value.PositionDict {
				base := allStat.PositionDict[posKey].Total
				positionDenominator := int64(base)
				if key == "all" {
					posIndex := parseIntDefault(posKey, -1)
					if posIndex >= 0 && posIndex < len(positionTotal) {
						positionDenominator = int64(positionTotal[posIndex])
					}
				}
				posStat.Proportion = newProportionStat(int64(posStat.Total), positionDenominator)
				if posStat.Total > 0 && base > 0 {
					posStat.Percent = rounded(float64(posStat.Total)/float64(base)*100, 2)
				} else {
					posStat.Percent = 0.0
				}
				posIndex := parseIntDefault(posKey, -1)
				if posIndex >= 0 && posIndex < len(positionTotal) && posStat.Total > 0 && positionTotal[posIndex] > 0 {
					posStat.PercentAll = rounded(float64(posStat.Total)/float64(positionTotal[posIndex])*100, 1)
				}
			}
		}
		for posKey, posStat := range allStat.PositionDict {
			posIndex := parseIntDefault(posKey, -1)
			if posIndex >= 0 && posIndex < len(positionTotal) && posStat.Total > 0 && positionTotal[posIndex] > 0 {
				posStat.Percent = rounded(float64(posStat.Total)/float64(positionTotal[posIndex])*100, 2)
			} else {
				posStat.Percent = 0.0
			}
		}
	}

	return &TuneStatsResponse{
		DataTotal:       logsTotal,
		SubstatDict:     substatDict,
		SubstatDistance: distances,
		SubstatPosTotal: substatPosTotal,
		PositionTotal:   positionTotal,
		Window:          window.Name,
	}, nil
}

func (a *App) posTotalExcludingEcho(e EchoLog, stats *TuneStatsResponse) int {
	if stats == nil {
		return 0
	}
	pos := currentPos(e)
	if pos < 0 || pos >= len(stats.PositionTotal) {
		return 0
	}
	total := stats.PositionTotal[pos]
	for _, substat := range []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5} {
		if substat == 0 {
			continue
		}
		idx := bitPos(substat)
		if idx >= 0 && idx < len(stats.SubstatPosTotal) {
			total -= stats.SubstatPosTotal[idx][pos]
		}
	}
	return total
}

func (a *App) fillCurrentPositionPercent(e EchoLog, stats *TuneStatsResponse) *TuneStatsResponse {
	out := cloneTuneStats(stats)
	if out == nil {
		return nil
	}
	pos := currentPos(e)
	posTotal := a.posTotalExcludingEcho(e, out)
	if posTotal <= 0 || pos >= len(twoCritPercent) {
		return out
	}
	for _, substat := range out.SubstatDict {
		show := ((e.SubstatAll >> substat.Number) & 1) == 0
		if show {
			substat.CurPosPercent = fmt.Sprintf("%.1f%%", rounded(float64(out.SubstatPosTotal[substat.Number][pos])*100/float64(posTotal), 1))
		} else {
			substat.CurPosPercent = ""
		}
		for _, value := range substat.ValueDict {
			posStat := value.PositionDict[strconv.Itoa(pos)]
			if show && posStat.Total > 0 {
				posStat.Percent = fmt.Sprintf("%.1f%%", rounded(float64(posStat.Total)*100/float64(posTotal), 1))
			} else if show {
				posStat.Percent = ""
			} else {
				posStat.Percent = ""
			}
		}
	}
	return out
}

func scoreEcho(e EchoLog, resonator, cost string) *EchoScore {
	if cost == "" {
		cost = "1C"
	}
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	score := &EchoScore{Name: template.Name, Resonator: template.Name}
	maxScore := template.EchoMaxScore[cost[:1]]
	if maxScore <= 0 {
		return score
	}
	fields := []*float64{&score.Substat1, &score.Substat2, &score.Substat3, &score.Substat4, &score.Substat5}
	substats := []int64{e.Substat1, e.Substat2, e.Substat3, e.Substat4, e.Substat5}
	total := 0.0
	for i, substat := range substats {
		if substat == 0 {
			continue
		}
		value := substatValueScore(substat, template)
		*fields[i] = rounded(value/maxScore*50, 2)
		total += *fields[i]
	}
	score.SubstatAll = rounded(template.MainstatMaxScore[cost]+total, 2)
	return score
}

func substatValueScore(substat int64, template resonatorTemplate) float64 {
	substatNum := bitPos(substat)
	if substatNum < 0 || substatNum >= len(substatDefs) {
		return 0
	}
	valueNum := bitPos(substat >> substatBitWidth)
	if valueNum < 0 || valueNum >= len(substatDefs[substatNum].Values) {
		return 0
	}
	def := substatDefs[substatNum]
	value := def.Values[valueNum].Value
	return template.SubstatWeight[def.NameCN] * value
}
