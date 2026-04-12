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
	OwnerUserID     int64  `json:"owner_user_id,omitempty"`
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

type PityAnalysisSummary struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type PityAnalysisMaxGapRow struct {
	Label       string `json:"label"`
	MaxGap      int    `json:"max_gap"`
	UserID      int64  `json:"user_id"`
	StartEchoID int64  `json:"start_echo_id"`
	EndEchoID   int64  `json:"end_echo_id"`
}

type PityAnalysisBucketRow struct {
	GapLabel     string  `json:"gap_label"`
	Trials       int64   `json:"trials"`
	Successes    int64   `json:"successes"`
	ActualRate   float64 `json:"actual_rate"`
	ExpectedRate float64 `json:"expected_rate"`
	DeltaRate    float64 `json:"delta_rate"`
	InternalRate float64 `json:"internal_rate"`
}

type PityAnalysisEvent struct {
	Key               string                  `json:"key"`
	Label             string                  `json:"label"`
	SuccessCount      int64                   `json:"success_count"`
	BaseRate          float64                 `json:"base_rate"`
	InternalBaseRate  float64                 `json:"internal_base_rate"`
	MaxInternalGap    int                     `json:"max_internal_gap"`
	MaxGapUserID      int64                   `json:"max_gap_user_id"`
	MaxGapStartEchoID int64                   `json:"max_gap_start_echo_id"`
	MaxGapEndEchoID   int64                   `json:"max_gap_end_echo_id"`
	HardPitySummary   string                  `json:"hard_pity_summary"`
	SoftPitySummary   string                  `json:"soft_pity_summary"`
	Buckets           []PityAnalysisBucketRow `json:"buckets"`
}

type PityAnalysisResponse struct {
	GeneratedAt          *time.Time                `json:"generated_at,omitempty"`
	EchoTotal            int64                     `json:"echo_total"`
	UserTotal            int64                     `json:"user_total"`
	MedianEchoesPerUser  float64                   `json:"median_echoes_per_user"`
	MaxEchoesPerUser     int64                     `json:"max_echoes_per_user"`
	TimeOrderMismatch    int64                     `json:"time_order_mismatch"`
	Summaries            []PityAnalysisSummary     `json:"summaries"`
	DefinitionNotes      []string                  `json:"definition_notes"`
	MethodNotes          []string                  `json:"method_notes"`
	Conclusions          []string                  `json:"conclusions"`
	SelectionBiasNotes   []string                  `json:"selection_bias_notes"`
	MaxGapRows           []PityAnalysisMaxGapRow   `json:"max_gap_rows"`
	Events               []PityAnalysisEvent       `json:"events"`
	StageSummaries       []PityAnalysisSummary     `json:"stage_summaries"`
	ContinuationRows     []PityContinuationRow     `json:"continuation_rows"`
	DoubleCritFutureRows []PityDoubleCritFutureRow `json:"double_crit_future_rows"`
	DoubleCritPathRows   []PityDoubleCritPathRow   `json:"double_crit_path_rows"`
}

type PityContinuationRow struct {
	StageOpened    int     `json:"stage_opened"`
	PrefixCategory string  `json:"prefix_category"`
	SampleCount    int64   `json:"sample_count"`
	ContinueCount  int64   `json:"continue_count"`
	StopCount      int64   `json:"stop_count"`
	ContinueRate   float64 `json:"continue_rate"`
}

type PityDoubleCritFutureRow struct {
	StageOpened          int     `json:"stage_opened"`
	PrefixCategory       string  `json:"prefix_category"`
	SampleCount          int64   `json:"sample_count"`
	CompletedCount       int64   `json:"completed_count"`
	CompletedRate        float64 `json:"completed_rate"`
	FinalDoubleCritCount int64   `json:"final_double_crit_count"`
	FinalDoubleCritRate  float64 `json:"final_double_crit_rate"`
}

type PityDoubleCritPathRow struct {
	PathLabel            string  `json:"path_label"`
	SampleCount          int64   `json:"sample_count"`
	EligibleCount        int64   `json:"eligible_count"`
	CompletedCount       int64   `json:"completed_count"`
	FinalDoubleCritCount int64   `json:"final_double_crit_count"`
	FinalDoubleCritRate  float64 `json:"final_double_crit_rate"`
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
