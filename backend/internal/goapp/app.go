package goapp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	substatBitWidth         = 13
	substatMask             = (1 << substatBitWidth) - 1
	authInvalidDetail       = "Token 无效或已过期"
	authForbiddenDetail     = "权限不足"
	authUnavailableDetail   = "鉴权服务不可用"
	tunerRecycledPerSubstat = 3
	expGold                 = 5000
)

type App struct {
	db          *pgxpool.Pool
	authURL     string
	httpClient  *http.Client
	ws          *wsManager
	statsMu     sync.RWMutex
	cachedStats *TuneStatsResponse
}

type contextKey string

const authInfoKey contextKey = "authInfo"

type AuthInfo struct {
	Permissions []string `json:"permissions"`
	OperatorID  *int64   `json:"operator_id"`
}

type EchoLog struct {
	ID         int64      `json:"id"`
	Substat1   int64      `json:"substat1"`
	Substat2   int64      `json:"substat2"`
	Substat3   int64      `json:"substat3"`
	Substat4   int64      `json:"substat4"`
	Substat5   int64      `json:"substat5"`
	SubstatAll int64      `json:"substat_all"`
	S1Desc     string     `json:"s1_desc"`
	S2Desc     string     `json:"s2_desc"`
	S3Desc     string     `json:"s3_desc"`
	S4Desc     string     `json:"s4_desc"`
	S5Desc     string     `json:"s5_desc"`
	Clazz      string     `json:"clazz"`
	UserID     int64      `json:"user_id"`
	OperatorID *int64     `json:"operator_id,omitempty"`
	Deleted    int        `json:"deleted"`
	TunedAt    *time.Time `json:"tuned_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type SubstatLog struct {
	ID         int64      `json:"id"`
	Substat    int        `json:"substat"`
	Value      int        `json:"value"`
	Position   int        `json:"position"`
	EchoID     int64      `json:"echo_id"`
	UserID     int64      `json:"user_id"`
	OperatorID *int64     `json:"operator_id,omitempty"`
	Timestamp  *time.Time `json:"timestamp,omitempty"`
	Deleted    int        `json:"deleted"`
}

type EchoTuneRequest struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Clazz      string `json:"clazz"`
	Substat1   int64  `json:"substat1"`
	Substat2   int64  `json:"substat2"`
	Substat3   int64  `json:"substat3"`
	Substat4   int64  `json:"substat4"`
	Substat5   int64  `json:"substat5"`
	SubstatAll int64  `json:"substat_all"`
	S1Desc     string `json:"s1_desc"`
	S2Desc     string `json:"s2_desc"`
	S3Desc     string `json:"s3_desc"`
	S4Desc     string `json:"s4_desc"`
	S5Desc     string `json:"s5_desc"`
	Position   int    `json:"position"`
	Substat    int    `json:"substat"`
	Value      int    `json:"value"`
}

type EchoFindRequest struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Clazz    string `json:"clazz"`
	Keyword  string `json:"keyword"`
	Substat1 int64  `json:"substat1"`
	Substat2 int64  `json:"substat2"`
	Substat3 int64  `json:"substat3"`
	Substat4 int64  `json:"substat4"`
	Substat5 int64  `json:"substat5"`
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PageResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	DataTotal int64       `json:"data_total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	PageTotal int         `json:"page_total"`
}

type SubstatValuePositionStat struct {
	Position   int             `json:"position"`
	Total      int             `json:"total"`
	Percent    any             `json:"percent"`
	PercentAll float64         `json:"percent_all,omitempty"`
	Proportion *ProportionStat `json:"proportion,omitempty"`
}

type SubstatValueStat struct {
	ValueNumber    int                                  `json:"value_number"`
	ValueDesc      string                               `json:"value_desc"`
	ValueDescFull  string                               `json:"value_desc_full"`
	Total          int                                  `json:"total"`
	Percent        any                                  `json:"percent"`
	PercentSubstat float64                              `json:"percent_substat,omitempty"`
	PositionDict   map[string]*SubstatValuePositionStat `json:"position_dict"`
	Proportion     *ProportionStat                      `json:"proportion,omitempty"`
}

type SubstatItem struct {
	Number        int                          `json:"number"`
	Name          string                       `json:"name"`
	NameCN        string                       `json:"name_cn"`
	Total         int                          `json:"total"`
	Percent       float64                      `json:"percent"`
	ValueDict     map[string]*SubstatValueStat `json:"value_dict"`
	CurPosPercent string                       `json:"cur_pos_percent"`
	Proportion    *ProportionStat              `json:"proportion,omitempty"`
}

type EchoScore struct {
	Name       string  `json:"name"`
	Resonator  string  `json:"resonator,omitempty"`
	Substat1   float64 `json:"substat1"`
	Substat2   float64 `json:"substat2"`
	Substat3   float64 `json:"substat3"`
	Substat4   float64 `json:"substat4"`
	Substat5   float64 `json:"substat5"`
	SubstatAll float64 `json:"substat_all"`
}

type TuneStatsResponse struct {
	DataTotal         int64                   `json:"data_total"`
	SubstatDict       map[string]*SubstatItem `json:"substat_dict"`
	SubstatDistance   []int                   `json:"substat_distance"`
	SubstatPosTotal   [][]int                 `json:"substat_pos_total"`
	PositionTotal     []int                   `json:"position_total"`
	Score             *EchoScore              `json:"score,omitempty"`
	ResonatorTemplate *resonatorTemplate      `json:"resonator_template,omitempty"`
	TwoCritPercent    float64                 `json:"two_crit_percent,omitempty"`
	Window            string                  `json:"window,omitempty"`
	BaselineCompare   map[string]any          `json:"baseline_compare,omitempty"`
}

type wsManager struct {
	mu          sync.RWMutex
	connections map[string]map[*wsConn]struct{}
}

type wsConn struct {
	conn net.Conn
	br   *bufio.Reader
	mu   sync.Mutex
}

type statsWindow struct {
	Name  string
	Limit int
	Days  int
}

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
	mux.HandleFunc("GET /", a.withPermission("view", a.handleRoot))
	mux.HandleFunc("GET /substat_logs", a.withPermission("view", a.handleListSubstatLogs))
	mux.HandleFunc("POST /tune_log/{id}/delete", a.withPermission("edit", a.handleDeleteTuneLogByID))
	mux.HandleFunc("POST /tune_log", a.withPermission("edit", a.handleAddTuneLog))
	mux.HandleFunc("GET /tune_stats", a.withPermission("view", a.handleTuneStats))
	mux.HandleFunc("GET /substat_distance_analysis", a.withPermission("view", a.handleSubstatDistanceAnalysis))
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
	return a.cors(mux)
}

func (a *App) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loadEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		_ = os.Setenv(key, value)
	}
}

func (a *App) withPermission(required string, next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info, err := a.validateRequest(r, required)
		if err != nil {
			a.writeAuthError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), authInfoKey, info)
		next(w, r.WithContext(ctx))
	}
}

func (a *App) validateRequest(r *http.Request, required string) (*AuthInfo, error) {
	token := extractToken(r)
	if token == "" {
		return nil, statusError{status: http.StatusUnauthorized, detail: authInvalidDetail}
	}
	payload := map[string]any{"token": token}
	if required != "" {
		payload["permission"] = required
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, a.authURL+"/api/validate", bytes.NewReader(body))
	if err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	defer resp.Body.Close()

	var result struct {
		Valid       bool     `json:"valid"`
		Reason      string   `json:"reason"`
		Permissions []string `json:"permissions"`
		ID          *int64   `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, statusError{status: http.StatusServiceUnavailable, detail: authUnavailableDetail}
	}
	if !result.Valid {
		if strings.EqualFold(result.Reason, "forbidden") {
			return nil, statusError{status: http.StatusForbidden, detail: authForbiddenDetail}
		}
		return nil, statusError{status: http.StatusUnauthorized, detail: authInvalidDetail}
	}
	if required != "" && !hasPermission(result.Permissions, required) {
		return nil, statusError{status: http.StatusForbidden, detail: authForbiddenDetail}
	}
	return &AuthInfo{Permissions: result.Permissions, OperatorID: result.ID}, nil
}

type statusError struct {
	status int
	detail string
}

func (e statusError) Error() string { return e.detail }

func (a *App) writeAuthError(w http.ResponseWriter, err error) {
	var se statusError
	if errors.As(err, &se) {
		writeJSONWithStatus(w, se.status, map[string]string{"detail": se.detail})
		return
	}
	writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
}

func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1])
		}
	}
	if token := strings.TrimSpace(r.Header.Get("X-Token")); token != "" {
		return token
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == "manage" || p == required {
			return true
		}
	}
	return false
}

func authInfoFromContext(ctx context.Context) *AuthInfo {
	info, _ := ctx.Value(authInfoKey).(*AuthInfo)
	return info
}

func operatorIDFromContext(ctx context.Context) *int64 {
	info := authInfoFromContext(ctx)
	if info == nil {
		return nil
	}
	return info.OperatorID
}

func canManage(ctx context.Context) bool {
	info := authInfoFromContext(ctx)
	if info == nil {
		return false
	}
	return hasPermission(info.Permissions, "manage")
}

func writeJSON(w http.ResponseWriter, payload any) {
	writeJSONWithStatus(w, http.StatusOK, payload)
}

func writeJSONWithStatus(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func success(message string, data any) SuccessResponse {
	return SuccessResponse{Code: 200, Message: message, Data: data}
}

func appError(message string, code int) ErrorResponse {
	return ErrorResponse{Code: code, Message: message}
}

func page(message string, data any, total int64, pageNum, pageSize int) PageResponse {
	pageTotal := 0
	if pageSize > 0 {
		pageTotal = int(total/int64(pageSize)) + 1
	}
	return PageResponse{Code: 200, Message: message, Data: data, DataTotal: total, Page: pageNum, PageSize: pageSize, PageTotal: pageTotal}
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func bitCount(bits int64) int {
	count := 0
	for bits != 0 {
		count++
		bits &= bits - 1
	}
	return count
}

func bitPos(bits int64) int {
	pos := 0
	for bits != 0 {
		if bits&1 == 1 {
			return pos
		}
		bits >>= 1
		pos++
	}
	return -1
}

func rounded(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Round(val*p) / p
}

func cloneTuneStats(src *TuneStatsResponse) *TuneStatsResponse {
	if src == nil {
		return nil
	}
	buf, _ := json.Marshal(src)
	var out TuneStatsResponse
	_ = json.Unmarshal(buf, &out)
	return &out
}

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
					posIndex, _ := strconv.Atoi(posKey)
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
				posIndex, _ := strconv.Atoi(posKey)
				if posIndex >= 0 && posIndex < len(positionTotal) && posStat.Total > 0 && positionTotal[posIndex] > 0 {
					posStat.PercentAll = rounded(float64(posStat.Total)/float64(positionTotal[posIndex])*100, 1)
				}
			}
		}
		for posKey, posStat := range allStat.PositionDict {
			posIndex, _ := strconv.Atoi(posKey)
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

func currentPos(e EchoLog) int {
	pos := 0
	if e.Substat1 != 0 {
		pos = 1
	}
	if e.Substat2 != 0 {
		pos = 2
	}
	if e.Substat3 != 0 {
		pos = 3
	}
	if e.Substat4 != 0 {
		pos = 4
	}
	return pos
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

func (wsm *wsManager) handle(operatorID string, conn *wsConn) {
	wsm.mu.Lock()
	if _, ok := wsm.connections[operatorID]; !ok {
		wsm.connections[operatorID] = map[*wsConn]struct{}{}
	}
	wsm.connections[operatorID][conn] = struct{}{}
	wsm.mu.Unlock()
	defer func() {
		wsm.mu.Lock()
		delete(wsm.connections[operatorID], conn)
		if len(wsm.connections[operatorID]) == 0 {
			delete(wsm.connections, operatorID)
		}
		wsm.mu.Unlock()
		_ = conn.conn.Close()
	}()
	for {
		if err := conn.readLoop(); err != nil {
			if err != io.EOF {
				log.Printf("websocket receive: %v", err)
			}
			return
		}
	}
}

func (wsm *wsManager) send(operatorID int64, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	for conn := range wsm.connections[strconv.FormatInt(operatorID, 10)] {
		_ = conn.writeText(string(body))
	}
}

func (a *App) scanEchoLog(row pgx.Row) (*EchoLog, error) {
	var item EchoLog
	if err := row.Scan(&item.ID, &item.Substat1, &item.Substat2, &item.Substat3, &item.Substat4, &item.Substat5, &item.SubstatAll, &item.S1Desc, &item.S2Desc, &item.S3Desc, &item.S4Desc, &item.S5Desc, &item.Clazz, &item.UserID, &item.OperatorID, &item.Deleted, &item.TunedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func (a *App) scanEchoLogs(rows pgx.Rows) ([]EchoLog, error) {
	var items []EchoLog
	for rows.Next() {
		item, err := a.scanEchoLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (a *App) handleProxyLogin(w http.ResponseWriter, r *http.Request) {
	a.proxyAuthRequest(w, r, http.MethodPost, "/api/login", r.Body)
}

func (a *App) handleProxyMe(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		writeJSONWithStatus(w, http.StatusUnauthorized, map[string]string{"detail": authInvalidDetail})
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.authURL+"/api/me", nil)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func (a *App) proxyAuthRequest(w http.ResponseWriter, r *http.Request, method, path string, body io.Reader) {
	req, err := http.NewRequestWithContext(r.Context(), method, a.authURL+path, body)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		writeJSONWithStatus(w, http.StatusServiceUnavailable, map[string]string{"detail": authUnavailableDetail})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	operatorID := strings.TrimSpace(r.URL.Query().Get("operator_id"))
	if operatorID == "" {
		writeJSON(w, appError("operator_id is required", 400))
		return
	}
	conn, err := upgradeWebsocket(w, r)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}
	a.ws.handle(operatorID, conn)
}

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"message": "Hello, Wuwa!"})
}

func (a *App) handleListSubstatLogs(w http.ResponseWriter, r *http.Request) {
	pageNum := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	offset := (pageNum - 1) * pageSize
	rows, err := a.db.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log order by id desc offset $1 limit $2", offset, pageSize)
	if err != nil {
		writeJSON(w, appError("failed to get tune logs", 500))
		return
	}
	defer rows.Close()
	var data []SubstatLog
	for rows.Next() {
		var item SubstatLog
		if err := rows.Scan(&item.ID, &item.Substat, &item.Value, &item.Position, &item.EchoID, &item.UserID, &item.OperatorID, &item.Timestamp, &item.Deleted); err != nil {
			writeJSON(w, appError("failed to get tune logs", 500))
			return
		}
		data = append(data, item)
	}
	var total int64
	if err := a.db.QueryRow(r.Context(), "select count(id) from wuwa_tune_log").Scan(&total); err != nil {
		writeJSON(w, appError("failed to get tune logs", 500))
		return
	}
	writeJSON(w, page("tune logs", data, total, pageNum, pageSize))
}

func (a *App) handleDeleteTuneLogByID(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	var tuneLog SubstatLog
	err := a.db.QueryRow(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where id = $1", id).Scan(&tuneLog.ID, &tuneLog.Substat, &tuneLog.Value, &tuneLog.Position, &tuneLog.EchoID, &tuneLog.UserID, &tuneLog.OperatorID, &tuneLog.Timestamp, &tuneLog.Deleted)
	if err != nil {
		writeJSON(w, appError("tune log not found", 500))
		return
	}
	if tuneLog.OperatorID == nil || (*tuneLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete this tune log", 403))
		return
	}
	var echoOperatorID *int64
	_ = a.db.QueryRow(r.Context(), "select operator_id from wuwa_echo_log where id = $1", tuneLog.EchoID).Scan(&echoOperatorID)
	if echoOperatorID != nil && *echoOperatorID != *operatorID && !isManager {
		writeJSON(w, appError("not authorized to delete tune log for this echo log", 403))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	defer tx.Rollback(r.Context())

	if tuneLog.Deleted == 0 {
		if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{tuneLog}, -1); err != nil {
			writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
			return
		}
	}
	tag, err := tx.Exec(r.Context(), "delete from wuwa_tune_log where id = $1", id)
	if err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError(fmt.Sprintf("failed to delete tune log %d", id), 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success(fmt.Sprintf("delete tune log %d", id), map[string]int64{"row_deleted": tag.RowsAffected()}))
}

func (a *App) handleAddTuneLog(w http.ResponseWriter, r *http.Request) {
	var payload SubstatLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	defer tx.Rollback(r.Context())

	var created SubstatLog
	err = tx.QueryRow(r.Context(), "insert into wuwa_tune_log (user_id, echo_id, position, substat, value, operator_id) values ($1, $2, $3, $4, $5, $6) returning id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted", payload.UserID, payload.EchoID, payload.Position, payload.Substat, payload.Value, *operatorID).Scan(&created.ID, &created.Substat, &created.Value, &created.Position, &created.EchoID, &created.UserID, &created.OperatorID, &created.Timestamp, &created.Deleted)
	if err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{created}, 1); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to add tune log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("add tune log", map[string]any{}))
}

func (a *App) handleTuneStats(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	afterID := parseInt64Default(r.URL.Query().Get("after_id"), 0)
	beforeID := parseInt64Default(r.URL.Query().Get("before_id"), 0)
	window := parseStatsWindow(r.URL.Query().Get("window"))

	var (
		stats *TuneStatsResponse
		err   error
	)
	if window.isAll() && size == 0 && afterID == 0 && beforeID == 0 {
		stats, err = a.loadTuneStatsFromAggregate(r.Context(), userID)
	}
	if stats == nil && err == nil {
		stats, err = a.computeTuneStats(r.Context(), size, userID, afterID, beforeID, window)
	}
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	if stats != nil {
		stats.Window = window.Name
	}
	if userID > 0 {
		var globalStats *TuneStatsResponse
		if window.isAll() && size == 0 && afterID == 0 && beforeID == 0 {
			globalStats, err = a.loadTuneStatsFromAggregate(r.Context(), 0)
		} else {
			globalStats, err = a.computeTuneStats(r.Context(), size, 0, afterID, beforeID, window)
		}
		if err != nil {
			writeJSON(w, appError("failed to get stats", 500))
			return
		}
		stats.BaselineCompare = buildTuneStatsBaselineCompare(stats, globalStats)
	}
	writeJSON(w, success("tune stats", stats))
}

func (a *App) handleSubstatDistanceAnalysis(w http.ResponseWriter, r *http.Request) {
	stats, err := a.computeTuneStats(r.Context(), parseIntDefault(r.URL.Query().Get("size"), 0), 0, 0, 0, parseStatsWindow(""))
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	writeJSON(w, success("tune stats", map[string]any{"data_total": stats.DataTotal, "substat_dict": stats.SubstatDict, "substat_distance": stats.SubstatDistance}))
}

func (a *App) handleAnalyzeEcho(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to get echo log", 500))
		return
	}
	stats := a.fillCurrentPositionPercent(payload, a.getCachedTuneStats())
	if stats == nil {
		stats = &TuneStatsResponse{SubstatDict: newSubstatDict(), PositionTotal: make([]int, 5), SubstatPosTotal: make([][]int, 13)}
	}
	resonator := r.URL.Query().Get("resonator")
	template, ok := resonatorTemplates[resonator]
	if !ok {
		template = defaultResonatorTemplate()
	}
	stats.ResonatorTemplate = &template
	stats.Score = scoreEcho(payload, resonator, r.URL.Query().Get("cost"))
	pos := currentPos(payload)
	critCount := bitCount(payload.SubstatAll & 0b11)
	if pos >= 0 && pos < len(twoCritPercent) && critCount >= 0 && critCount < len(twoCritPercent[pos]) {
		stats.TwoCritPercent = twoCritPercent[pos][critCount]
	}
	writeJSON(w, success("echo log", stats))
}

func (a *App) handleEchoDcrit(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	beforeID := parseInt64Default(r.URL.Query().Get("before_id"), 0)
	afterID := parseInt64Default(r.URL.Query().Get("after_id"), 0)
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	window := parseStatsWindow(r.URL.Query().Get("window"))
	if window.isAll() && size == 0 && beforeID == 0 && afterID == 0 {
		if data, err := a.loadEchoDcritFromAggregate(r.Context(), userID); err == nil && data != nil {
			if userID > 0 {
				if globalData, globalErr := a.loadEchoDcritFromAggregate(r.Context(), 0); globalErr == nil && globalData != nil {
					data["baseline_compare"] = map[string]any{
						"dcrit_rate": buildRateComparison(
							data["dcrit_rate_stats"].(*ProportionStat),
							globalData["dcrit_rate_stats"].(*ProportionStat),
						),
					}
				}
			}
			data["window"] = window.Name
			writeJSON(w, success("test", data))
			return
		}
	}
	query := "select id, substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	var args []any
	arg := 1
	if userID > 0 {
		query += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, userID)
		arg++
	}
	if afterID > 0 {
		query += fmt.Sprintf(" and id > $%d", arg)
		args = append(args, afterID)
		arg++
	}
	if beforeID > 0 {
		query += fmt.Sprintf(" and id < $%d", arg)
		args = append(args, beforeID)
		arg++
	}
	if since := window.sinceTime(); since != nil {
		query += fmt.Sprintf(" and updated_at >= $%d", arg)
		args = append(args, *since)
		arg++
	}
	if effectiveSize := window.applyLimit(size); effectiveSize > 0 {
		query += fmt.Sprintf(" limit %d", effectiveSize)
	}
	rows, err := a.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to test", 500))
		return
	}
	defer rows.Close()
	echoCount := 0
	dcritTotal := 0
	counts := map[string]map[string]int{}
	for rows.Next() {
		echoCount++
		var id, s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&id, &s1, &s2, &s3, &s4, &s5); err != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		substatAll := s1 | s2 | s3 | s4 | s5
		if substatAll&0b11 == 0b11 {
			dcritTotal++
			critRateNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b01)
			critDmgNum := firstTierForMask([]int64{s1, s2, s3, s4, s5}, 0b10)
			rk := strconv.Itoa(bitPos(critRateNum))
			dk := strconv.Itoa(bitPos(critDmgNum))
			if _, ok := counts[rk]; !ok {
				counts[rk] = map[string]int{}
			}
			counts[rk][dk]++
		}
	}
	resp := map[string]any{
		"echo_count":       echoCount,
		"dcrit_total":      dcritTotal,
		"counts":           counts,
		"dcrit_rate_stats": newProportionStat(int64(dcritTotal), int64(echoCount)),
		"window":           window.Name,
	}
	if userID > 0 {
		globalResp, globalErr := a.computeEchoDcritRaw(r.Context(), 0, size, afterID, beforeID, window)
		if globalErr != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		resp["baseline_compare"] = map[string]any{
			"dcrit_rate": buildRateComparison(resp["dcrit_rate_stats"].(*ProportionStat), globalResp["dcrit_rate_stats"].(*ProportionStat)),
		}
	}
	writeJSON(w, success("test", resp))
}

func firstTierForMask(substats []int64, mask int64) int64 {
	for _, substat := range substats {
		if substat&mask != 0 {
			return substat >> substatBitWidth
		}
	}
	return 0
}

func (a *App) handleTestZero(w http.ResponseWriter, r *http.Request) {
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	query := "select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0"
	if size > 0 {
		query += fmt.Sprintf(" limit %d", size)
	}
	rows, err := a.db.Query(r.Context(), query)
	if err != nil {
		writeJSON(w, appError("failed to test", 500))
		return
	}
	defer rows.Close()
	echoCount, dcritTotal, dcrit2, dcrit3, dcrit4 := 0, 0, 0, 0, 0
	for rows.Next() {
		var s1, s2, s3, s4, s5 int64
		if err := rows.Scan(&s1, &s2, &s3, &s4, &s5); err != nil {
			writeJSON(w, appError("failed to test", 500))
			return
		}
		echoCount++
		substatAll := s1 | s2 | s3 | s4 | s5
		if substatAll&0b11 == 0b11 {
			dcritTotal++
			if (s1|s2)&0b11 == 0b11 {
				dcrit2++
			}
			if (s1|s2|s3)&0b11 == 0b11 {
				dcrit3++
			}
			if (s1|s2|s3|s4)&0b11 == 0b11 {
				dcrit4++
			}
		}
	}
	rate := func(v int) string {
		if echoCount == 0 {
			return "0%"
		}
		return fmt.Sprintf("%v%%", float64(v)/float64(echoCount)*100)
	}
	perEcho := func(v int) string {
		if echoCount == 0 || v == 0 {
			return "0"
		}
		return fmt.Sprintf("%v", 1.0/(float64(v)/float64(echoCount)))
	}
	writeJSON(w, success("test", map[string]any{
		"echo_count":        echoCount,
		"dcrit_total":       dcritTotal,
		"dcrit2_total":      dcrit2,
		"dcrit3_total":      dcrit3,
		"dcrit4_total":      dcrit4,
		"dcrit2_rate":       rate(dcrit2),
		"dcrit3_rate":       rate(dcrit3),
		"dcrit4_rate":       rate(dcrit4),
		"dcrit2_per_echoes": perEcho(dcrit2),
		"dcrit3_per_echoes": perEcho(dcrit3),
		"dcrit4_per_echoes": perEcho(dcrit4),
	}))
}

func (a *App) handlePredictEchoSubstat(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	target := payload.SubstatAll & substatMask
	if target < 0 {
		writeJSON(w, appError("substat_target must >= 0", 500))
		return
	}
	if bitPos(target) >= 5 {
		writeJSON(w, success("predict echo substat", map[string]any{"count_total": 0, "count": make([]int, 14), "percent": make([]float64, 14)}))
		return
	}
	rows, err := a.db.Query(r.Context(), "select substat1, substat2, substat3, substat4, substat5 from wuwa_echo_log where deleted = 0 and (substat_all & $1) = $1", target)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	defer rows.Close()
	counts := make([]int, 15)
	for rows.Next() {
		var log EchoLog
		if err := rows.Scan(&log.Substat1, &log.Substat2, &log.Substat3, &log.Substat4, &log.Substat5); err != nil {
			writeJSON(w, appError("failed to get echo logs", 500))
			return
		}
		switch {
		case payload.Substat1 == 0:
			incrementPredictCount(counts, log.Substat1&substatMask)
		case payload.Substat1&substatMask != log.Substat1&substatMask:
			continue
		case payload.Substat2 == 0:
			incrementPredictCount(counts, log.Substat2&substatMask)
		case payload.Substat2&substatMask != log.Substat2&substatMask:
			continue
		case payload.Substat3 == 0:
			incrementPredictCount(counts, log.Substat3&substatMask)
		case payload.Substat3&substatMask != log.Substat3&substatMask:
			continue
		case payload.Substat4 == 0:
			incrementPredictCount(counts, log.Substat4&substatMask)
		case payload.Substat4&substatMask != log.Substat4&substatMask:
			continue
		case payload.Substat5 == 0:
			incrementPredictCount(counts, log.Substat5&substatMask)
		}
	}
	counts = counts[:13]
	total := 0
	for _, count := range counts {
		total += count
	}
	percent := make([]float64, len(counts))
	for i, count := range counts {
		if total > 0 {
			percent[i] = rounded(float64(count)/float64(total)*100, 1)
		}
	}
	writeJSON(w, success("predict echo substat", map[string]any{"count_total": total, "count": counts, "percent": percent}))
}

func (a *App) handleWriteEchoSubstatAll(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all from wuwa_echo_log where deleted = 0")
	if err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	defer rows.Close()
	var total, successTotal int64
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	defer tx.Rollback(r.Context())
	for rows.Next() {
		total++
		var id, s1, s2, s3, s4, s5, substatAll int64
		if err := rows.Scan(&id, &s1, &s2, &s3, &s4, &s5, &substatAll); err != nil {
			writeJSON(w, appError("failed to write substat all", 500))
			return
		}
		if substatAll == 0 {
			calculated := (s1 | s2 | s3 | s4 | s5) & substatMask
			if _, err := tx.Exec(r.Context(), "update wuwa_echo_log set substat_all = $1 where id = $2", calculated, id); err != nil {
				writeJSON(w, appError("failed to write substat all", 500))
				return
			}
			successTotal++
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to write substat all", 500))
		return
	}
	writeJSON(w, success("write substat all", map[string]int64{"success_total": successTotal, "total": total}))
}

func (a *App) handleWriteSubstatUserID(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "select id, echo_id, user_id from wuwa_tune_log where deleted = 0 order by id desc")
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	defer rows.Close()
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	defer tx.Rollback(r.Context())
	var total, successTotal int64
	for rows.Next() {
		total++
		var id, echoID, userID int64
		if err := rows.Scan(&id, &echoID, &userID); err != nil {
			writeJSON(w, appError("failed to get stats", 500))
			return
		}
		if echoID > 0 && userID == 0 {
			var echoUserID int64
			if err := tx.QueryRow(r.Context(), "select user_id from wuwa_echo_log where id = $1", echoID).Scan(&echoUserID); err == nil && echoUserID > 0 {
				if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set user_id = $1 where id = $2", echoUserID, id); err != nil {
					writeJSON(w, appError("failed to get stats", 500))
					return
				}
				successTotal++
			}
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to get stats", 500))
		return
	}
	writeJSON(w, success("write id", map[string]int64{"success_total": successTotal, "total": total}))
}

func (a *App) handleListEchoLogs(w http.ResponseWriter, r *http.Request) {
	pageNum := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	offset := (pageNum - 1) * pageSize
	rows, err := a.db.Query(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log order by updated_at desc offset $1 limit $2", offset, pageSize)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	var total int64
	if err := a.db.QueryRow(r.Context(), "select count(id) from wuwa_echo_log where deleted = 0").Scan(&total); err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	writeJSON(w, page("echo logs", items, total, pageNum, pageSize))
}

func applyEchoChanges(existing *EchoLog, payload EchoLog) {
	if payload.Substat1 != 0 {
		existing.Substat1 = payload.Substat1
	}
	if payload.Substat2 != 0 {
		existing.Substat2 = payload.Substat2
	}
	if payload.Substat3 != 0 {
		existing.Substat3 = payload.Substat3
	}
	if payload.Substat4 != 0 {
		existing.Substat4 = payload.Substat4
	}
	if payload.Substat5 != 0 {
		existing.Substat5 = payload.Substat5
	}
	if payload.SubstatAll != 0 {
		existing.SubstatAll = payload.SubstatAll
	}
	if payload.S1Desc != "" {
		existing.S1Desc = payload.S1Desc
	}
	if payload.S2Desc != "" {
		existing.S2Desc = payload.S2Desc
	}
	if payload.S3Desc != "" {
		existing.S3Desc = payload.S3Desc
	}
	if payload.S4Desc != "" {
		existing.S4Desc = payload.S4Desc
	}
	if payload.S5Desc != "" {
		existing.S5Desc = payload.S5Desc
	}
	if payload.Clazz != "" {
		existing.Clazz = payload.Clazz
	}
	if payload.UserID != 0 {
		existing.UserID = payload.UserID
	}
	now := time.Now()
	existing.UpdatedAt = &now
}

func replaceEchoChanges(existing *EchoLog, payload EchoLog) {
	existing.Substat1 = payload.Substat1
	existing.Substat2 = payload.Substat2
	existing.Substat3 = payload.Substat3
	existing.Substat4 = payload.Substat4
	existing.Substat5 = payload.Substat5
	existing.SubstatAll = payload.SubstatAll
	existing.S1Desc = payload.S1Desc
	existing.S2Desc = payload.S2Desc
	existing.S3Desc = payload.S3Desc
	existing.S4Desc = payload.S4Desc
	existing.S5Desc = payload.S5Desc
	existing.Clazz = payload.Clazz
	existing.UserID = payload.UserID
	now := time.Now()
	existing.UpdatedAt = &now
}

func (a *App) handleCreateEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	now := time.Now()
	if payload.TunedAt == nil {
		payload.TunedAt = &now
	}
	payload.CreatedAt = &now
	payload.UpdatedAt = &now
	payload.OperatorID = operatorID
	row := a.db.QueryRow(r.Context(), "insert into wuwa_echo_log (substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, tuned_at, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", payload.Substat1, payload.Substat2, payload.Substat3, payload.Substat4, payload.Substat5, payload.SubstatAll, payload.S1Desc, payload.S2Desc, payload.S3Desc, payload.S4Desc, payload.S5Desc, payload.Clazz, payload.UserID, *operatorID, payload.TunedAt, payload.CreatedAt, payload.UpdatedAt)
	created, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), a.db, nil, created); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), a.db, nil, created); err != nil {
		writeJSON(w, appError("failed to create echo log", 500))
		return
	}
	a.ws.send(*operatorID, map[string]any{"type": "create_echo_log", "data": created})
	writeJSON(w, success("create echo log", created))
}

func (a *App) handleUpdateEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoLog
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	existing, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", payload.ID))
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	if existing.OperatorID == nil || (*existing.OperatorID != *operatorID && !canManage(r.Context())) {
		writeJSON(w, appError("not authorized to update this echo log", 403))
		return
	}
	beforeEcho := *existing
	replaceEchoChanges(existing, payload)
	row := a.db.QueryRow(r.Context(), "update wuwa_echo_log set substat1=$1, substat2=$2, substat3=$3, substat4=$4, substat5=$5, substat_all=$6, s1_desc=$7, s2_desc=$8, s3_desc=$9, s4_desc=$10, s5_desc=$11, clazz=$12, user_id=$13, updated_at=$14 where id=$15 returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", existing.Substat1, existing.Substat2, existing.Substat3, existing.Substat4, existing.Substat5, existing.SubstatAll, existing.S1Desc, existing.S2Desc, existing.S3Desc, existing.S4Desc, existing.S5Desc, existing.Clazz, existing.UserID, existing.UpdatedAt, existing.ID)
	updated, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), a.db, &beforeEcho, updated); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), a.db, &beforeEcho, updated); err != nil {
		writeJSON(w, appError("failed to update echo log", 500))
		return
	}
	a.ws.send(*updated.OperatorID, map[string]any{"type": "update_echo_log", "data": updated})
	writeJSON(w, success("update echo log", updated))
}

func (a *App) handleTuneEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoTuneRequest
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())

	var echoLog *EchoLog
	now := time.Now()
	if payload.ID > 0 {
		row := tx.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", payload.ID)
		echoLog, err = a.scanEchoLog(row)
		if err != nil {
			writeJSON(w, appError("echo log not found", 404))
			return
		}
		if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
			writeJSON(w, appError("not authorized to tune this echo log", 403))
			return
		}
	} else {
		if payload.UserID == 0 {
			writeJSON(w, appError("user_id is required", 400))
			return
		}
		if payload.Clazz == "" {
			writeJSON(w, appError("clazz is required", 400))
			return
		}
		row := tx.QueryRow(r.Context(), "insert into wuwa_echo_log (user_id, clazz, tuned_at, created_at, updated_at, operator_id) values ($1,$2,$3,$4,$5,$6) returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", payload.UserID, payload.Clazz, now, now, now, *operatorID)
		echoLog, err = a.scanEchoLog(row)
		if err != nil {
			writeJSON(w, appError("failed to tune echo log", 500))
			return
		}
	}
	beforeEcho := *echoLog
	applyEchoChanges(echoLog, EchoLog{Substat1: payload.Substat1, Substat2: payload.Substat2, Substat3: payload.Substat3, Substat4: payload.Substat4, Substat5: payload.Substat5, SubstatAll: payload.SubstatAll, S1Desc: payload.S1Desc, S2Desc: payload.S2Desc, S3Desc: payload.S3Desc, S4Desc: payload.S4Desc, S5Desc: payload.S5Desc, Clazz: payload.Clazz, UserID: payload.UserID})
	row := tx.QueryRow(r.Context(), "update wuwa_echo_log set substat1=$1, substat2=$2, substat3=$3, substat4=$4, substat5=$5, substat_all=$6, s1_desc=$7, s2_desc=$8, s3_desc=$9, s4_desc=$10, s5_desc=$11, clazz=$12, user_id=$13, updated_at=$14 where id=$15 returning id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at", echoLog.Substat1, echoLog.Substat2, echoLog.Substat3, echoLog.Substat4, echoLog.Substat5, echoLog.SubstatAll, echoLog.S1Desc, echoLog.S2Desc, echoLog.S3Desc, echoLog.S4Desc, echoLog.S5Desc, echoLog.Clazz, echoLog.UserID, echoLog.UpdatedAt, echoLog.ID)
	echoLog, err = a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	var tuneLog SubstatLog
	err = tx.QueryRow(r.Context(), "insert into wuwa_tune_log (user_id, echo_id, position, substat, value, operator_id) values ($1,$2,$3,$4,$5,$6) returning id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted", echoLog.UserID, echoLog.ID, payload.Position, payload.Substat, payload.Value, *operatorID).Scan(&tuneLog.ID, &tuneLog.Substat, &tuneLog.Value, &tuneLog.Position, &tuneLog.EchoID, &tuneLog.UserID, &tuneLog.OperatorID, &tuneLog.Timestamp, &tuneLog.Deleted)
	if err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, []SubstatLog{tuneLog}, 1); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, echoLog); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, echoLog); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to tune echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	a.ws.send(*operatorID, map[string]any{"type": "tune_echo_log", "data": map[string]any{"echo_log": echoLog, "tune_log": tuneLog}})
	writeJSON(w, success("tune echo log", map[string]any{"echo_log": echoLog, "tune_log": tuneLog}))
}

func (a *App) handleDeleteEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id))
	if err != nil {
		writeJSON(w, appError("echo log not found", 500))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete this echo log", 403))
		return
	}
	emptyEcho := echoLog.Substat1 == 0 && echoLog.Substat2 == 0 && echoLog.Substat3 == 0 && echoLog.Substat4 == 0 && echoLog.Substat5 == 0
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and deleted = 0", id)
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	result := map[string]any{}
	if emptyEcho {
		if _, err := tx.Exec(r.Context(), "delete from wuwa_tune_log where echo_id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if _, err := tx.Exec(r.Context(), "delete from wuwa_echo_log where id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		result = map[string]any{"deleted": "hard", "id": id}
	} else {
		beforeEcho := *echoLog
		afterEcho := beforeEcho
		afterEcho.Deleted = 1
		tag1, err := tx.Exec(r.Context(), "update wuwa_echo_log set deleted = 1 where id = $1", id)
		if err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set deleted = 1 where echo_id = $1", id); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
			writeJSON(w, appError("failed to delete echo log", 500))
			return
		}
		result = map[string]any{"rows_affected": tag1.RowsAffected()}
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, -1); err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to delete echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	a.ws.send(*echoLog.OperatorID, map[string]any{"type": "delete_echo_log", "data": map[string]any{"id": id, "deleted": map[bool]string{true: "hard", false: "soft"}[emptyEcho]}})
	writeJSON(w, success("delete echo log", result))
}

func (a *App) handleRecoverEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id))
	if err != nil {
		writeJSON(w, appError("echo log not found", 500))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to recover this echo log", 403))
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and deleted = 1", id)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	beforeEcho := *echoLog
	afterEcho := beforeEcho
	afterEcho.Deleted = 0
	tag, err := tx.Exec(r.Context(), "update wuwa_echo_log set deleted = 0 where id = $1", id)
	if err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if _, err := tx.Exec(r.Context(), "update wuwa_tune_log set deleted = 0 where echo_id = $1", id); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, 1); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyEchoDcritDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := a.applyEchoSummaryDelta(r.Context(), tx, &beforeEcho, &afterEcho); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to recover echo log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("recover echo log", map[string]int64{"rows_affected": tag.RowsAffected()}))
}

func (a *App) handleGetEchoLog(w http.ResponseWriter, r *http.Request) {
	id := parseInt64Default(r.PathValue("id"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	var row pgx.Row
	if id > 0 {
		row = a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", id)
	} else {
		row = a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0 and operator_id = $1 order by updated_at desc limit 1", *operatorID)
	}
	echoLog, err := a.scanEchoLog(row)
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	writeJSON(w, success("echo log", map[string]any{
		"id":          echoLog.ID,
		"substat1":    echoLog.Substat1,
		"substat2":    echoLog.Substat2,
		"substat3":    echoLog.Substat3,
		"substat4":    echoLog.Substat4,
		"substat5":    echoLog.Substat5,
		"substat_all": echoLog.SubstatAll,
		"s1_desc":     echoLog.S1Desc,
		"s2_desc":     echoLog.S2Desc,
		"s3_desc":     echoLog.S3Desc,
		"s4_desc":     echoLog.S4Desc,
		"s5_desc":     echoLog.S5Desc,
		"clazz":       echoLog.Clazz,
		"user_id":     echoLog.UserID,
		"operator_id": echoLog.OperatorID,
		"deleted":     echoLog.Deleted,
		"tuned_at":    echoLog.TunedAt,
		"created_at":  echoLog.CreatedAt,
		"updated_at":  echoLog.UpdatedAt,
		"pos_total":   a.posTotalExcludingEcho(*echoLog, a.getCachedTuneStats()),
	}))
}

func (a *App) handleFindEchoLog(w http.ResponseWriter, r *http.Request) {
	var payload EchoFindRequest
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	hasSubstatFilter := (payload.Substat1 | payload.Substat2 | payload.Substat3 | payload.Substat4 | payload.Substat5) != 0
	keyword := strings.TrimSpace(payload.Keyword)
	if !hasSubstatFilter && payload.ID <= 0 && payload.UserID <= 0 && payload.Clazz == "" && keyword == "" {
		writeJSON(w, success("no search condition specified, return empty list", []EchoLog{}))
		return
	}
	pageSize := parseIntDefault(r.URL.Query().Get("page_size"), 20)
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0"
	var args []any
	arg := 1
	if payload.ID > 0 {
		query += fmt.Sprintf(" and id = $%d", arg)
		args = append(args, payload.ID)
		arg++
	}
	for _, filter := range []struct {
		column string
		bits   int64
	}{
		{"substat1", payload.Substat1},
		{"substat2", payload.Substat2},
		{"substat3", payload.Substat3},
		{"substat4", payload.Substat4},
		{"substat5", payload.Substat5},
	} {
		if filter.bits == 0 {
			continue
		}
		if filter.bits&^int64(substatMask) == 0 {
			query += fmt.Sprintf(" and (%s & $%d) = $%d", filter.column, arg, arg)
		} else {
			query += fmt.Sprintf(" and %s = $%d", filter.column, arg)
		}
		args = append(args, filter.bits)
		arg++
	}
	if payload.UserID > 0 {
		query += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, payload.UserID)
		arg++
	}
	if payload.Clazz != "" {
		query += fmt.Sprintf(" and clazz = $%d", arg)
		args = append(args, payload.Clazz)
		arg++
	}
	if keyword != "" {
		query += fmt.Sprintf(" and (clazz ilike $%d or s1_desc ilike $%d or s2_desc ilike $%d or s3_desc ilike $%d or s4_desc ilike $%d or s5_desc ilike $%d or cast(user_id as text) ilike $%d or cast(id as text) ilike $%d)", arg, arg, arg, arg, arg, arg, arg, arg)
		args = append(args, "%"+keyword+"%")
		arg++
	}
	query += fmt.Sprintf(" order by updated_at desc limit %d", pageSize)
	rows, err := a.db.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		writeJSON(w, appError("failed to find echo logs", 500))
		return
	}
	writeJSON(w, success("find echo logs", items))
}

func (a *App) handleDeleteSubstatByEchoPos(w http.ResponseWriter, r *http.Request) {
	echoID := parseInt64Default(r.PathValue("echoId"), 0)
	pos := parseIntDefault(r.PathValue("pos"), 0)
	operatorID := operatorIDFromContext(r.Context())
	if operatorID == nil {
		writeJSON(w, appError("operator not found", 401))
		return
	}
	isManager := canManage(r.Context())
	echoLog, err := a.scanEchoLog(a.db.QueryRow(r.Context(), "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where id = $1", echoID))
	if err != nil {
		writeJSON(w, appError("echo log not found", 404))
		return
	}
	if echoLog.OperatorID == nil || (*echoLog.OperatorID != *operatorID && !isManager) {
		writeJSON(w, appError("not authorized to delete substats for this echo log", 403))
		return
	}
	query := "update wuwa_tune_log set deleted = 1 where echo_id = $1 and position = $2"
	args := []any{echoID, pos}
	if !isManager {
		query += " and operator_id = $3"
		args = append(args, *operatorID)
	}
	selectQuery := "select id, substat, value, position, echo_id, user_id, operator_id, timestamp, deleted from wuwa_tune_log where echo_id = $1 and position = $2 and deleted = 0"
	selectArgs := []any{echoID, pos}
	if !isManager {
		selectQuery += " and operator_id = $3"
		selectArgs = append(selectArgs, *operatorID)
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	defer tx.Rollback(r.Context())
	affectedRows, err := tx.Query(r.Context(), selectQuery, selectArgs...)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	affectedLogs, err := collectTuneLogs(affectedRows)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	tag, err := tx.Exec(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	if err := a.applyTuneStatsDelta(r.Context(), tx, affectedLogs, -1); err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, appError("failed to delete substat log", 500))
		return
	}
	_ = a.refreshCachedTuneStats(r.Context())
	writeJSON(w, success("delete substat log", map[string]int64{"rows_affected": tag.RowsAffected()}))
}

func (a *App) handleEchoLogsAnalysis(w http.ResponseWriter, r *http.Request) {
	userID := parseInt64Default(r.URL.Query().Get("user_id"), 0)
	size := parseIntDefault(r.URL.Query().Get("size"), 0)
	targetBits := parseInt64Default(r.URL.Query().Get("target_bits"), 0b11)
	substatSinceDate := strings.TrimSpace(r.URL.Query().Get("substat_since_date"))
	window := parseStatsWindow(r.URL.Query().Get("window"))
	if window.isAll() && size == 0 && substatSinceDate == "" {
		if data, err := a.loadEchoSummaryFromAggregate(r.Context(), userID, targetBits); err == nil && data != nil {
			if userID > 0 {
				if globalData, globalErr := a.loadEchoSummaryFromAggregate(r.Context(), 0, targetBits); globalErr == nil && globalData != nil {
					data["baseline_compare"] = map[string]any{
						"target_rate": buildRateComparison(
							data["target_rate_stats"].(*ProportionStat),
							globalData["target_rate_stats"].(*ProportionStat),
						),
					}
				}
			}
			data["window"] = window.Name
			writeJSON(w, success("echo logs analysis", data))
			return
		}
	}
	effectiveSize := window.applyLimit(size)
	items, total, err := a.loadEchoLogsAnalysisItems(r.Context(), userID, effectiveSize, targetBits, window, substatSinceDate)
	if err != nil {
		writeJSON(w, appError("failed to get echo logs", 500))
		return
	}
	resp := computeEchoLogsAnalysisFromItems(items, total, targetBits)
	resp["window"] = window.Name
	if userID > 0 {
		globalItems, globalTotal, globalErr := a.loadEchoLogsAnalysisItems(r.Context(), 0, effectiveSize, targetBits, window, substatSinceDate)
		if globalErr != nil {
			writeJSON(w, appError("failed to get echo logs", 500))
			return
		}
		globalResp := computeEchoLogsAnalysisFromItems(globalItems, globalTotal, targetBits)
		resp["baseline_compare"] = map[string]any{
			"target_rate": buildRateComparison(resp["target_rate_stats"].(*ProportionStat), globalResp["target_rate_stats"].(*ProportionStat)),
		}
	}
	writeJSON(w, success("echo logs analysis", resp))
}

func (a *App) loadEchoLogsAnalysisItems(ctx context.Context, userID int64, effectiveSize int, targetBits int64, window statsWindow, substatSinceDate string) ([]EchoLog, int64, error) {
	selectSQL := "select id, substat1, substat2, substat3, substat4, substat5, substat_all, s1_desc, s2_desc, s3_desc, s4_desc, s5_desc, clazz, user_id, operator_id, deleted, tuned_at, created_at, updated_at from wuwa_echo_log where deleted = 0"
	countSQL := "select count(id) from wuwa_echo_log where deleted = 0"
	var args []any
	arg := 1
	if userID > 0 {
		selectSQL += fmt.Sprintf(" and user_id = $%d", arg)
		countSQL += fmt.Sprintf(" and user_id = $%d", arg)
		args = append(args, userID)
		arg++
		if substatSinceDate != "" {
			parsed, err := time.Parse("2006-01-02", substatSinceDate)
			if err == nil {
				startAt := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 4, 0, 0, 0, parsed.Location())
				rows, err := a.db.Query(ctx, "select echo_id from wuwa_tune_log where deleted = 0 and user_id = $1 and timestamp >= $2", userID, startAt)
				if err == nil {
					defer rows.Close()
					idsMap := map[int64]struct{}{}
					var ids []int64
					for rows.Next() {
						var echoID int64
						if rows.Scan(&echoID) == nil {
							if _, ok := idsMap[echoID]; !ok {
								idsMap[echoID] = struct{}{}
								ids = append(ids, echoID)
							}
						}
					}
					if len(ids) > 0 {
						selectSQL += fmt.Sprintf(" and id = any($%d)", arg)
						countSQL += fmt.Sprintf(" and id = any($%d)", arg)
						args = append(args, ids)
						arg++
					} else {
						selectSQL += " and id = -1"
						countSQL += " and id = -1"
					}
				}
			}
		}
	}
	if since := window.sinceTime(); since != nil {
		selectSQL += fmt.Sprintf(" and updated_at >= $%d", arg)
		countSQL += fmt.Sprintf(" and updated_at >= $%d", arg)
		args = append(args, *since)
		arg++
	}
	selectSQL += " order by updated_at desc"
	if effectiveSize > 0 {
		selectSQL += fmt.Sprintf(" limit %d", effectiveSize)
	}
	rows, err := a.db.Query(ctx, selectSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := a.scanEchoLogs(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := a.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if effectiveSize > 0 {
		total = int64(len(items))
	}
	return items, total, nil
}

func parseIntDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	if value, err := strconv.Atoi(raw); err == nil {
		return value
	}
	return fallback
}

func parseInt64Default(raw string, fallback int64) int64 {
	if raw == "" {
		return fallback
	}
	if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return value
	}
	return fallback
}

func parseStatsWindow(raw string) statsWindow {
	switch strings.TrimSpace(raw) {
	case "", "all":
		return statsWindow{Name: "all"}
	case "last_100":
		return statsWindow{Name: "last_100", Limit: 100}
	case "last_500":
		return statsWindow{Name: "last_500", Limit: 500}
	case "last_1000":
		return statsWindow{Name: "last_1000", Limit: 1000}
	case "day_7":
		return statsWindow{Name: "day_7", Days: 7}
	case "day_30":
		return statsWindow{Name: "day_30", Days: 30}
	default:
		return statsWindow{Name: "all"}
	}
}

func (w statsWindow) isAll() bool {
	return w.Name == "" || w.Name == "all"
}

func (w statsWindow) applyLimit(size int) int {
	if w.Limit > 0 && size > 0 {
		if size < w.Limit {
			return size
		}
		return w.Limit
	}
	if w.Limit > 0 {
		return w.Limit
	}
	return size
}

func (w statsWindow) sinceTime() *time.Time {
	if w.Days <= 0 {
		return nil
	}
	t := time.Now().AddDate(0, 0, -w.Days)
	return &t
}

func computeEchoLogsAnalysisFromItems(items []EchoLog, total int64, targetBits int64) map[string]any {
	found := false
	idx := 0
	targetCount := 0
	targetEchoDistance := -1
	targetSubstatDistance := -1
	substatTotal := 0
	tunerRecycled := 0
	expTotal := 0
	expRecycled := 0
	for _, echoLog := range items {
		substatAll := (echoLog.Substat1 | echoLog.Substat2 | echoLog.Substat3 | echoLog.Substat4 | echoLog.Substat5) & substatMask
		substatCount := bitCount(substatAll)
		substatTotal += substatCount
		expTotal += expTable[0][substatCount]
		if substatAll&targetBits == targetBits {
			targetCount++
			if !found {
				found = true
				targetEchoDistance = idx
				targetSubstatDistance = substatTotal
			}
		} else {
			tunerRecycled += substatCount * tunerRecycledPerSubstat
			expRecycled += expReturn[substatCount]
		}
		idx++
	}
	if !found {
		targetEchoDistance = idx
		targetSubstatDistance = substatTotal
	}
	tunerConsumed := int(math.Ceil(float64(substatTotal*10 - tunerRecycled)))
	expConsumed := int(math.Ceil(float64(expTotal-expRecycled) / expGold))
	resp := map[string]any{
		"sample_size":             total,
		"target_echo_distance":    targetEchoDistance,
		"target_substat_distance": targetSubstatDistance,
		"target":                  targetCount,
		"target_avg_echo":         0.0,
		"target_avg_substat":      0.0,
		"tuner_consumed":          tunerConsumed,
		"tuner_consumed_avg":      0.0,
		"exp_consumed":            expConsumed,
		"exp_consumed_avg":        0.0,
		"target_rate_stats":       newProportionStat(int64(targetCount), total),
	}
	if targetCount > 0 {
		resp["target_avg_echo"] = rounded(float64(total)/float64(targetCount), 1)
		resp["target_avg_substat"] = rounded(float64(substatTotal)/float64(targetCount), 1)
		resp["tuner_consumed_avg"] = int(math.Ceil(float64(tunerConsumed) / float64(targetCount)))
		resp["exp_consumed_avg"] = int(math.Ceil(float64(expConsumed) / float64(targetCount)))
	}
	return resp
}

func incrementPredictCount(counts []int, bits int64) {
	index := bitPos(bits)
	if index >= 0 && index < len(counts) {
		counts[index]++
	}
}

func commandTagRowsAffected(tag pgconn.CommandTag) int64 {
	return tag.RowsAffected()
}

func upgradeWebsocket(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	if !headerContainsToken(r.Header, "Connection", "Upgrade") || !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return nil, errors.New("not websocket upgrade")
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, errors.New("missing websocket key")
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("hijack unsupported")
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}
	accept := websocketAccept(key)
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := buf.WriteString(response); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := buf.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &wsConn{conn: conn, br: bufio.NewReader(conn)}, nil
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerContainsToken(header http.Header, key, token string) bool {
	for _, value := range header.Values(key) {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

func (c *wsConn) writeText(message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	payload := []byte(message)
	header := []byte{0x81}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	case len(payload) <= 65535:
		header = append(header, 126, byte(len(payload)>>8), byte(len(payload)))
	default:
		header = append(header, 127,
			byte(uint64(len(payload))>>56), byte(uint64(len(payload))>>48), byte(uint64(len(payload))>>40), byte(uint64(len(payload))>>32),
			byte(uint64(len(payload))>>24), byte(uint64(len(payload))>>16), byte(uint64(len(payload))>>8), byte(uint64(len(payload))))
	}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}

func (c *wsConn) readLoop() error {
	first, err := c.br.ReadByte()
	if err != nil {
		return err
	}
	second, err := c.br.ReadByte()
	if err != nil {
		return err
	}
	opcode := first & 0x0f
	masked := second&0x80 != 0
	payloadLen := int(second & 0x7f)
	switch payloadLen {
	case 126:
		b1, err := c.br.ReadByte()
		if err != nil {
			return err
		}
		b2, err := c.br.ReadByte()
		if err != nil {
			return err
		}
		payloadLen = int(b1)<<8 | int(b2)
	case 127:
		var size uint64
		for i := 0; i < 8; i++ {
			b, err := c.br.ReadByte()
			if err != nil {
				return err
			}
			size = (size << 8) | uint64(b)
		}
		payloadLen = int(size)
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(c.br, mask[:]); err != nil {
			return err
		}
	}
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(c.br, payload); err != nil {
		return err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	switch opcode {
	case 0x8:
		return io.EOF
	case 0x9:
		return c.writeControl(0xA, payload)
	default:
		return nil
	}
}

func (c *wsConn) writeControl(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	header := []byte{0x80 | opcode, byte(len(payload))}
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err := c.conn.Write(payload)
	return err
}
