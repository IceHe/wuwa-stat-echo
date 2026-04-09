package goapp

import (
	"context"

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
