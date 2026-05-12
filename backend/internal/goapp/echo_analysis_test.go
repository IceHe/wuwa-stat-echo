package goapp

import (
	"math"
	"net/http/httptest"
	"testing"
)

func TestIsUsableEchoLogsAnalysisSummary(t *testing.T) {
	valid := map[string]any{
		"sample_size": int64(6),
		"target":      int64(6),
	}
	if !isUsableEchoLogsAnalysisSummary(valid) {
		t.Fatalf("expected valid aggregate summary to be accepted")
	}

	invalid := map[string]any{
		"sample_size": int64(5),
		"target":      int64(6),
	}
	if isUsableEchoLogsAnalysisSummary(invalid) {
		t.Fatalf("expected invalid aggregate summary to be rejected")
	}
}

func TestNewProportionStatHandlesInvalidCounts(t *testing.T) {
	stat := newProportionStat(6, 5)
	if stat == nil {
		t.Fatalf("expected stat")
	}
	if math.IsNaN(stat.Rate) || math.IsInf(stat.Rate, 0) {
		t.Fatalf("expected finite rate, got %v", stat.Rate)
	}
	if math.IsNaN(stat.CI95Low) || math.IsNaN(stat.CI95High) {
		t.Fatalf("expected finite confidence interval values, got low=%v high=%v", stat.CI95Low, stat.CI95High)
	}
	if stat.CI95Low != 0 || stat.CI95High != 0 {
		t.Fatalf("expected confidence interval to stay unset for invalid counts, got low=%v high=%v", stat.CI95Low, stat.CI95High)
	}
}

func TestWriteJSONWithStatusFallsBackOnMarshalError(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeJSONWithStatus(recorder, 200, map[string]any{
		"bad": math.NaN(),
	})
	if recorder.Code != 500 {
		t.Fatalf("expected fallback status 500, got %d", recorder.Code)
	}
	if recorder.Body.Len() == 0 {
		t.Fatalf("expected fallback body")
	}
}
