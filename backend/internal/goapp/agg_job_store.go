package goapp

import (
	"context"
	"net/http"
	"time"
)

type aggRebuildJob struct {
	ID         int64      `json:"id"`
	JobType    string     `json:"job_type"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Message    string     `json:"message"`
}

func (a *App) ensureAggRebuildJobsReady(ctx context.Context) error {
	if _, err := a.db.Exec(ctx, `
		create table if not exists agg_rebuild_jobs (
			id bigserial primary key,
			job_type text not null,
			status text not null,
			started_at timestamptz,
			finished_at timestamptz,
			message text not null default ''
		)
	`); err != nil {
		return err
	}
	return nil
}

func (a *App) createAggRebuildJob(ctx context.Context, jobType string) (*aggRebuildJob, error) {
	var job aggRebuildJob
	err := a.db.QueryRow(ctx, `
		insert into agg_rebuild_jobs (job_type, status, started_at, message)
		values ($1, 'running', now(), '')
		returning id, job_type, status, started_at, finished_at, message
	`, jobType).Scan(&job.ID, &job.JobType, &job.Status, &job.StartedAt, &job.FinishedAt, &job.Message)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (a *App) finishAggRebuildJob(ctx context.Context, jobID int64, status string, message string) (*aggRebuildJob, error) {
	var job aggRebuildJob
	err := a.db.QueryRow(ctx, `
		update agg_rebuild_jobs
		set status = $2, finished_at = now(), message = $3
		where id = $1
		returning id, job_type, status, started_at, finished_at, message
	`, jobID, status, message).Scan(&job.ID, &job.JobType, &job.Status, &job.StartedAt, &job.FinishedAt, &job.Message)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (a *App) loadAggRebuildJob(ctx context.Context, jobID int64) (*aggRebuildJob, error) {
	var job aggRebuildJob
	err := a.db.QueryRow(ctx, `
		select id, job_type, status, started_at, finished_at, message
		from agg_rebuild_jobs
		where id = $1
	`, jobID).Scan(&job.ID, &job.JobType, &job.Status, &job.StartedAt, &job.FinishedAt, &job.Message)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (a *App) runAggRebuildJob(ctx context.Context, jobType string, fn func(context.Context) error) (*aggRebuildJob, error) {
	job, err := a.createAggRebuildJob(ctx, jobType)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx); err != nil {
		failedJob, finishErr := a.finishAggRebuildJob(ctx, job.ID, "failed", err.Error())
		if finishErr == nil {
			job = failedJob
		}
		return job, err
	}
	finishedJob, err := a.finishAggRebuildJob(ctx, job.ID, "success", "ok")
	if err != nil {
		return job, err
	}
	return finishedJob, nil
}

func (a *App) handleGetAggRebuildJob(w http.ResponseWriter, r *http.Request) {
	jobID := parseInt64Default(r.PathValue("jobID"), 0)
	job, err := a.loadAggRebuildJob(r.Context(), jobID)
	if err != nil {
		writeJSON(w, appError("agg rebuild job not found", 404))
		return
	}
	writeJSON(w, success("agg rebuild job", job))
}
