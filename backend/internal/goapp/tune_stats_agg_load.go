package goapp

import (
	"context"
	"fmt"
)

func (a *App) loadTuneStatsFromAggregate(ctx context.Context, userID int64) (*TuneStatsResponse, error) {
	rows, err := a.db.Query(ctx, `
		select substat, value, position, count
		from agg_tune_substat_counts
		where bucket_type = $1 and bucket_key = $2 and user_id = $3
		order by substat, value, position
	`, aggBucketTypeAll, aggBucketKeyAll, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	substatDict := newSubstatDict()
	positionTotal := make([]int, 5)
	substatPosTotal := make([][]int, len(substatDefs))
	for i := range substatPosTotal {
		substatPosTotal[i] = make([]int, 5)
	}

	var dataTotal int64
	rowCount := 0
	for rows.Next() {
		rowCount++
		var substat, value, position int
		var count int64
		if err := rows.Scan(&substat, &value, &position, &count); err != nil {
			return nil, err
		}

		if substat < 0 || substat >= len(substatDefs) || position < 0 || position >= len(positionTotal) {
			continue
		}

		item := substatDict[fmt.Sprint(substat)]
		if item == nil {
			continue
		}

		if value == aggValueAll {
			item.Total += int(count)
			item.ValueDict["all"].Total += int(count)
			item.ValueDict["all"].PositionDict[fmt.Sprint(position)].Total += int(count)
			positionTotal[position] += int(count)
			substatPosTotal[substat][position] += int(count)
			dataTotal += count
			continue
		}

		valueItem := item.ValueDict[fmt.Sprint(value)]
		if valueItem == nil {
			continue
		}
		valueItem.Total += int(count)
		valueItem.PositionDict[fmt.Sprint(position)].Total += int(count)
	}

	if rowCount == 0 {
		return nil, nil
	}

	for _, item := range substatDict {
		allValue := item.ValueDict["all"]
		item.Proportion = newProportionStat(int64(item.Total), dataTotal)
		if dataTotal > 0 {
			item.Percent = rounded(float64(item.Total)/float64(dataTotal)*100, 2)
		}
		for key, valueItem := range item.ValueDict {
			denominator := int64(allValue.Total)
			if key == "all" {
				denominator = dataTotal
			}
			valueItem.Proportion = newProportionStat(int64(valueItem.Total), denominator)
			if item.Total > 0 {
				valueItem.PercentSubstat = rounded(float64(valueItem.Total)/float64(item.Total)*100, 2)
			}
			if key == "all" {
				if dataTotal > 0 {
					valueItem.Percent = rounded(float64(valueItem.Total)/float64(dataTotal)*100, 2)
				}
			} else if allValue.Total > 0 {
				valueItem.Percent = rounded(float64(valueItem.Total)/float64(allValue.Total)*100, 2)
			}

			for posKey, positionItem := range valueItem.PositionDict {
				allAtPos := allValue.PositionDict[posKey].Total
				positionDenominator := int64(allAtPos)
				if key == "all" {
					posIndex := parseIntDefault(posKey, -1)
					if posIndex >= 0 && posIndex < len(positionTotal) {
						positionDenominator = int64(positionTotal[posIndex])
					}
				}
				positionItem.Proportion = newProportionStat(int64(positionItem.Total), positionDenominator)
				if positionItem.Total > 0 && allAtPos > 0 {
					positionItem.Percent = rounded(float64(positionItem.Total)/float64(allAtPos)*100, 2)
				} else {
					positionItem.Percent = 0.0
				}
				posIndex := parseIntDefault(posKey, -1)
				if posIndex >= 0 && posIndex < len(positionTotal) && positionItem.Total > 0 && positionTotal[posIndex] > 0 {
					positionItem.PercentAll = rounded(float64(positionItem.Total)/float64(positionTotal[posIndex])*100, 1)
				}
			}
		}

		for posKey, positionItem := range allValue.PositionDict {
			posIndex := parseIntDefault(posKey, -1)
			if posIndex >= 0 && posIndex < len(positionTotal) && positionItem.Total > 0 && positionTotal[posIndex] > 0 {
				positionItem.Percent = rounded(float64(positionItem.Total)/float64(positionTotal[posIndex])*100, 2)
			} else {
				positionItem.Percent = 0.0
			}
		}
	}

	distances, err := a.computeTuneStatDistances(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &TuneStatsResponse{
		DataTotal:       dataTotal,
		SubstatDict:     substatDict,
		SubstatDistance: distances,
		SubstatPosTotal: substatPosTotal,
		PositionTotal:   positionTotal,
	}, nil
}

func (a *App) computeTuneStatDistances(ctx context.Context, userID int64) ([]int, error) {
	query := `select substat from wuwa_tune_log where deleted = 0`
	args := make([]any, 0, 1)
	if userID > 0 {
		query += ` and user_id = $1`
		args = append(args, userID)
	}
	query += ` order by id desc`

	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distances := make([]int, len(substatDefs))
	for i := range distances {
		distances[i] = -1
	}

	index := -1
	remaining := len(substatDefs)
	for rows.Next() {
		index++
		var substat int
		if err := rows.Scan(&substat); err != nil {
			return nil, err
		}
		if substat < 0 || substat >= len(distances) {
			continue
		}
		if distances[substat] == -1 {
			distances[substat] = index
			remaining--
			if remaining == 0 {
				break
			}
		}
	}

	return distances, rows.Err()
}
