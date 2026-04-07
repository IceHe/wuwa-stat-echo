package goapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	aggBucketTypeAll = "all"
	aggBucketKeyAll  = "all"
	aggValueAll      = -1
)

type dbExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (a *App) ensureTuneStatsAggregateReady(ctx context.Context) error {
	if _, err := a.db.Exec(ctx, `
		create table if not exists agg_tune_substat_counts (
			bucket_type text not null,
			bucket_key text not null,
			user_id bigint not null,
			substat integer not null,
			value integer not null,
			position integer not null,
			count bigint not null default 0,
			updated_at timestamptz not null default now(),
			primary key (bucket_type, bucket_key, user_id, substat, value, position)
		)
	`); err != nil {
		return err
	}

	var aggCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from agg_tune_substat_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll).Scan(&aggCount); err != nil {
		return err
	}
	if aggCount > 0 {
		return nil
	}

	var rawCount int64
	if err := a.db.QueryRow(ctx, `select count(*) from wuwa_tune_log where deleted = 0`).Scan(&rawCount); err != nil {
		return err
	}
	if rawCount == 0 {
		return nil
	}

	return a.rebuildTuneStatsAggregate(ctx)
}

func (a *App) rebuildTuneStatsAggregate(ctx context.Context) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from agg_tune_substat_counts where bucket_type = $1 and bucket_key = $2`, aggBucketTypeAll, aggBucketKeyAll); err != nil {
		return fmt.Errorf("clear agg_tune_substat_counts: %w", err)
	}

	insertStatements := []struct {
		sql  string
		args []any
	}{
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, user_id, substat, value, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by user_id, substat, value, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, user_id, substat, $3, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by user_id, substat, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll, aggValueAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, 0, substat, value, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by substat, value, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll},
		},
		{
			sql: `
			insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count)
			select $1, $2, 0, substat, $3, position, count(*)
			from wuwa_tune_log
			where deleted = 0
			group by substat, position
			`,
			args: []any{aggBucketTypeAll, aggBucketKeyAll, aggValueAll},
		},
	}

	for index, statement := range insertStatements {
		if _, err := tx.Exec(ctx, statement.sql, statement.args...); err != nil {
			return fmt.Errorf("rebuild agg_tune_substat_counts step %d: %w", index+1, err)
		}
	}

	return tx.Commit(ctx)
}

func (a *App) handleRebuildTuneStatsAggregate(w http.ResponseWriter, r *http.Request) {
	job, err := a.runAggRebuildJob(r.Context(), "rebuild_tune_stats", func(ctx context.Context) error {
		if err := a.rebuildTuneStatsAggregate(ctx); err != nil {
			return err
		}
		return a.refreshCachedTuneStats(ctx)
	})
	if err != nil {
		writeJSONWithStatus(w, http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "failed to rebuild tune stats aggregate",
			"data":    map[string]any{"job": job},
		})
		return
	}
	writeJSON(w, success("rebuild tune stats aggregate", map[string]any{
		"bucket_type": aggBucketTypeAll,
		"bucket_key":  aggBucketKeyAll,
		"job":         job,
	}))
}

func (a *App) applyTuneStatsDelta(ctx context.Context, q dbExecutor, logs []SubstatLog, delta int64) error {
	if len(logs) == 0 || delta == 0 {
		return nil
	}

	statement := `
		insert into agg_tune_substat_counts (bucket_type, bucket_key, user_id, substat, value, position, count, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, now())
		on conflict (bucket_type, bucket_key, user_id, substat, value, position)
		do update set count = agg_tune_substat_counts.count + excluded.count, updated_at = now()
	`

	for _, logItem := range logs {
		if logItem.Substat < 0 || logItem.Substat >= len(substatDefs) {
			continue
		}
		if logItem.Position < 0 || logItem.Position >= 5 {
			continue
		}

		rows := [][4]int64{
			{logItem.UserID, int64(logItem.Substat), int64(logItem.Value), int64(logItem.Position)},
			{logItem.UserID, int64(logItem.Substat), aggValueAll, int64(logItem.Position)},
			{0, int64(logItem.Substat), int64(logItem.Value), int64(logItem.Position)},
			{0, int64(logItem.Substat), aggValueAll, int64(logItem.Position)},
		}
		for _, row := range rows {
			if _, err := q.Exec(ctx, statement, aggBucketTypeAll, aggBucketKeyAll, row[0], row[1], row[2], row[3], delta); err != nil {
				return err
			}
		}
	}

	return nil
}

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

func collectTuneLogs(rows pgx.Rows) ([]SubstatLog, error) {
	defer rows.Close()
	var logs []SubstatLog
	for rows.Next() {
		var item SubstatLog
		if err := rows.Scan(&item.ID, &item.Substat, &item.Value, &item.Position, &item.EchoID, &item.UserID, &item.OperatorID, &item.Timestamp, &item.Deleted); err != nil {
			return nil, err
		}
		logs = append(logs, item)
	}
	return logs, rows.Err()
}
