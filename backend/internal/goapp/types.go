package goapp

import (
	"bufio"
	"net"
	"net/http"
	"sync"
	"time"

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
	db                 *pgxpool.Pool
	authURL            string
	httpClient         *http.Client
	ws                 *wsManager
	statsMu            sync.RWMutex
	cachedStats        *TuneStatsResponse
	maxGapMu           sync.RWMutex
	substatMaxGapCache map[int64]*SubstatMaxGapResponse
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

type ScoreTemplateSyncRequest struct {
	Field     string `json:"field"`
	Value     string `json:"value"`
	Resonator string `json:"resonator"`
	Cost      string `json:"cost"`
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

type SubstatMaxGapRow struct {
	Substat         int    `json:"substat"`
	Name            string `json:"name"`
	NameCN          string `json:"name_cn"`
	MaxGap          int    `json:"max_gap"`
	OccurrenceCount int    `json:"occurrence_count"`
	LeadingGap      int    `json:"leading_gap"`
	TrailingGap     int    `json:"trailing_gap"`
	MaxGapStartID   int64  `json:"max_gap_start_id"`
	MaxGapEndID     int64  `json:"max_gap_end_id"`
}

type SubstatMaxGapResponse struct {
	UserID              int64              `json:"user_id"`
	ScopeLabel          string             `json:"scope_label"`
	TuneLogTotal        int                `json:"tune_log_total"`
	GeneratedAt         *time.Time         `json:"generated_at,omitempty"`
	LastForcedRefreshAt *time.Time         `json:"last_forced_refresh_at,omitempty"`
	RefreshAvailableAt  *time.Time         `json:"refresh_available_at,omitempty"`
	CacheHit            bool               `json:"cache_hit"`
	ForceApplied        bool               `json:"force_applied"`
	RefreshBlocked      bool               `json:"refresh_blocked"`
	Rows                []SubstatMaxGapRow `json:"rows"`
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
