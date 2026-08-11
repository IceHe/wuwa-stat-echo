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

func TestDaniyaTemplateScoreMatchesReferenceTotal(t *testing.T) {
	echo := EchoLog{
		Substat1: (1 << 2) | (1 << (4 + substatBitWidth)),  // 攻击 9.4%
		Substat2: (1 << 6) | (1 << (2 + substatBitWidth)),  // 防御固定值 60
		Substat3: (1 << 12) | (1 << (3 + substatBitWidth)), // 共鸣解放 8.6%
		Substat4: (1 << 0) | (1 << (3 + substatBitWidth)),  // 暴击 8.1%
		Substat5: (1 << 1) | (1 << (1 + substatBitWidth)),  // 暴击伤害 13.8%
	}

	score := scoreEcho(echo, "达妮娅", "3C属伤")
	if score.Resonator != "达妮娅" {
		t.Fatalf("expected 达妮娅 template, got %q", score.Resonator)
	}
	if math.Abs(score.SubstatAll-35.81) > 0.0001 {
		t.Fatalf("expected raw-sum total 35.81, got %.2f", score.SubstatAll)
	}
	if rounded(score.Substat1+score.Substat2+score.Substat3+score.Substat4+score.Substat5+6.85, 2) != 35.82 {
		t.Fatalf("fixture must distinguish raw total from displayed-component sum")
	}
}

func TestDaniyaTemplateConfigAndUnknownFallback(t *testing.T) {
	daniya, ok := resonatorTemplates["达妮娅"]
	if !ok {
		t.Fatalf("expected 达妮娅 score template")
	}
	if daniya.EchoMaxScore["4"] != 85.939 || daniya.EchoMaxScore["3"] != 83.88 || daniya.SubstatWeight["攻击"] != 1.2 || daniya.SubstatWeight["共鸣解放"] != 0.85 {
		t.Fatalf("unexpected 达妮娅 template: %+v", daniya)
	}

	rebecca, ok := resonatorTemplates["丽贝卡"]
	if !ok {
		t.Fatalf("expected 丽贝卡 score template")
	}
	if rebecca.EchoMaxScore["4"] != 85.475 || rebecca.SubstatWeight["普攻"] != 0.81 || rebecca.SubstatWeight["共鸣效率"] != 0.25 {
		t.Fatalf("unexpected 丽贝卡 template: %+v", rebecca)
	}

	lucy, ok := resonatorTemplates["露西"]
	if !ok {
		t.Fatalf("expected 露西 score template")
	}
	if lucy.EchoMaxScore["4"] != 86.518 || lucy.SubstatWeight["重击"] != 0.9 || lucy.MainstatMaxScore["1C"] != 6.44 {
		t.Fatalf("unexpected 露西 template: %+v", lucy)
	}

	rocciaFrost, ok := resonatorTemplates["校长霜渐"]
	if !ok {
		t.Fatalf("expected 校长霜渐 score template")
	}
	if rocciaFrost.EchoMaxScore["4"] != 87.71 || rocciaFrost.SubstatWeight["普攻"] != 0.9 || rocciaFrost.MainstatMaxScore["4C"] != 8.4 {
		t.Fatalf("unexpected 校长霜渐 template: %+v", rocciaFrost)
	}

	rocciaEcho, ok := resonatorTemplates["校长声骸"]
	if !ok {
		t.Fatalf("expected 校长声骸 score template")
	}
	if rocciaEcho.EchoMaxScore["4"] != 74.609 || rocciaEcho.SubstatWeight["攻击"] != 1.3 || rocciaEcho.SubstatWeight["共鸣技能"] != 0.12 {
		t.Fatalf("unexpected 校长声骸 template: %+v", rocciaEcho)
	}

	chisa, ok := resonatorTemplates["千咲"]
	if !ok {
		t.Fatalf("expected 千咲 score template")
	}
	if chisa.EchoMaxScore["1"] != 79.977 || chisa.SubstatWeight["共鸣效率"] != 0.25 || chisa.SubstatWeight["共鸣解放"] != 0.605 {
		t.Fatalf("unexpected 千咲 template: %+v", chisa)
	}

	sui, ok := resonatorTemplates["穗穗"]
	if !ok {
		t.Fatalf("expected 穗穗 score template")
	}
	if sui.EchoMaxScore["4"] != 54.427 || sui.MainstatMaxScore["3C属伤"] != 13.13 || sui.SubstatWeight["生命"] != 1.2 || sui.SubstatWeight["共鸣效率"] != 1 || sui.SubstatWeight["攻击"] != 0 || sui.SubstatWeight["攻击固定值"] != 0 {
		t.Fatalf("unexpected 穗穗 template: %+v", sui)
	}

	score := scoreEcho(EchoLog{}, "未配置角色", "3C属伤")
	if score.Resonator != "通用" {
		t.Fatalf("expected unknown resonator to fall back to 通用, got %q", score.Resonator)
	}
}
