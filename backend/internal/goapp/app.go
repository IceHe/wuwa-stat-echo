package goapp

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New() (*App, error) {
	loadEnv(filepath.Join(".", ".env"))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://127.0.0.1:8080"
	}
	timeoutSeconds := 3.0
	if raw := os.Getenv("AUTH_SERVICE_TIMEOUT_SECONDS"); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	app := &App{
		db:         pool,
		authURL:    strings.TrimRight(authURL, "/"),
		httpClient: &http.Client{Timeout: time.Duration(timeoutSeconds * float64(time.Second))},
		ws: &wsManager{
			connections: map[string]map[*wsConn]struct{}{},
		},
		substatMaxGapCache: map[int64]*SubstatMaxGapResponse{},
	}

	if err := app.ensureTuneStatsAggregateReady(context.Background()); err != nil {
		log.Printf("init tune stats aggregate failed: %v", err)
	}
	if err := app.ensureAggRebuildJobsReady(context.Background()); err != nil {
		log.Printf("init agg rebuild jobs failed: %v", err)
	}
	if err := app.ensureEchoDcritAggregateReady(context.Background()); err != nil {
		log.Printf("init echo dcrit aggregate failed: %v", err)
	}
	if err := app.ensureEchoSummaryAggregateReady(context.Background()); err != nil {
		log.Printf("init echo summary aggregate failed: %v", err)
	}
	if err := app.refreshCachedTuneStats(context.Background()); err != nil {
		log.Printf("init tune stats failed: %v", err)
	}
	return app, nil
}

func (a *App) Close() {
	a.db.Close()
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /ws", a.handleWebsocket)
	mux.HandleFunc("POST /auth/login", a.handleProxyLogin)
	mux.HandleFunc("GET /auth/me", a.handleProxyMe)
	mux.HandleFunc("GET /substat_logs", a.withPermission("view", a.handleListSubstatLogs))
	mux.HandleFunc("POST /tune_log/{id}/delete", a.withPermission("edit", a.handleDeleteTuneLogByID))
	mux.HandleFunc("POST /tune_log", a.withPermission("edit", a.handleAddTuneLog))
	mux.HandleFunc("GET /tune_stats", a.withPermission("view", a.handleTuneStats))
	mux.HandleFunc("GET /substat_distance_analysis", a.withPermission("view", a.handleSubstatDistanceAnalysis))
	mux.HandleFunc("GET /stats/substat_max_gap", a.withPermission("view", a.handleSubstatMaxGap))
	mux.HandleFunc("POST /analyze_echo", a.withPermission("view", a.handleAnalyzeEcho))
	mux.HandleFunc("GET /counts/echo_dcrit", a.withPermission("view", a.handleEchoDcrit))
	mux.HandleFunc("GET /test/0", a.withPermission("view", a.handleTestZero))
	mux.HandleFunc("POST /predict/echo_substat", a.withPermission("view", a.handlePredictEchoSubstat))
	mux.HandleFunc("POST /decision/echo-next-step", a.withPermission("view", a.handleDecisionEchoNextStep))
	mux.HandleFunc("POST /simulator/echo-future", a.withPermission("view", a.handleSimulatorEchoFuture))
	mux.HandleFunc("POST /simulator/echo-compare", a.withPermission("view", a.handleSimulatorEchoCompare))
	mux.HandleFunc("POST /admin/stats/rebuild/tune", a.withPermission("manage", a.handleRebuildTuneStatsAggregate))
	mux.HandleFunc("POST /admin/stats/rebuild/dcrit", a.withPermission("manage", a.handleRebuildEchoDcritAggregate))
	mux.HandleFunc("POST /admin/stats/rebuild/echo_summary", a.withPermission("manage", a.handleRebuildEchoSummaryAggregate))
	mux.HandleFunc("GET /admin/stats/rebuild/{jobID}", a.withPermission("manage", a.handleGetAggRebuildJob))
	mux.HandleFunc("POST /admin/stats/reconcile", a.withPermission("manage", a.handleReconcileAggregates))
	mux.HandleFunc("GET /db/echo_logs/write_substat_all", a.withPermission("manage", a.handleWriteEchoSubstatAll))
	mux.HandleFunc("GET /db/substat_logs/write_user_id", a.withPermission("manage", a.handleWriteSubstatUserID))
	mux.HandleFunc("GET /echo_logs", a.withPermission("view", a.handleListEchoLogs))
	mux.HandleFunc("POST /echo_log", a.withPermission("edit", a.handleCreateEchoLog))
	mux.HandleFunc("PATCH /echo_log", a.withPermission("edit", a.handleUpdateEchoLog))
	mux.HandleFunc("POST /echo_log/tune", a.withPermission("edit", a.handleTuneEchoLog))
	mux.HandleFunc("DELETE /echo_log/{id}", a.withPermission("edit", a.handleDeleteEchoLog))
	mux.HandleFunc("POST /echo_log/{id}/recover", a.withPermission("edit", a.handleRecoverEchoLog))
	mux.HandleFunc("GET /echo_log/{id}", a.withPermission("view", a.handleGetEchoLog))
	mux.HandleFunc("POST /echo_log/find", a.withPermission("view", a.handleFindEchoLog))
	mux.HandleFunc("DELETE /echo_log/{echoId}/substat_pos/{pos}", a.withPermission("edit", a.handleDeleteSubstatByEchoPos))
	mux.HandleFunc("GET /echo_logs/analysis", a.withPermission("view", a.handleEchoLogsAnalysis))
	mux.HandleFunc("POST /viewer/score_template_sync", a.withPermission("view", a.handleViewerScoreTemplateSync))
	mux.HandleFunc("GET /score_templates", a.withPermission("view", a.handleGetScoreTemplates))
	mux.HandleFunc("GET /", a.withPermission("view", a.handleRoot))
	return a.cors(mux)
}
