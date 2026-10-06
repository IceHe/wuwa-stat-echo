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

	score := scoreEcho(echo, "达妮娅", "3C属伤", "")
	if score.Resonator != "达妮娅" {
		t.Fatalf("expected 达妮娅 template, got %q", score.Resonator)
	}
	if math.Abs(score.SubstatAll-28.94) > 0.0001 {
		t.Fatalf("expected XW-UID-compatible substat total 28.94, got %.2f", score.SubstatAll)
	}
}

func TestDaniyaTemplateConfigAndUnknownFallback(t *testing.T) {
	daniya, ok := resonatorTemplates["达妮娅"]
	if !ok {
		t.Fatalf("expected 达妮娅 score template")
	}
	if daniya.EchoMaxScore["4"] != 85.939 || daniya.EchoMaxScore["3"] != 83.88 || daniya.SubstatWeight["攻击"] != 0.11 || daniya.SubstatWeight["技能伤害加成"] != 0.85 {
		t.Fatalf("unexpected 达妮娅 template: %+v", daniya)
	}

	rebecca, ok := resonatorTemplates["丽贝卡"]
	if !ok {
		t.Fatalf("expected 丽贝卡 score template")
	}
	if rebecca.EchoMaxScore["4"] != 86.518 || rebecca.SubstatWeight["技能伤害加成"] != 0.9 || rebecca.SubstatWeight["共鸣效率"] != 0.25 || rebecca.SubstatWeight["攻击固定值"] != 0.11 {
		t.Fatalf("unexpected 丽贝卡 template: %+v", rebecca)
	}

	lucy, ok := resonatorTemplates["露西"]
	if !ok {
		t.Fatalf("expected 露西 score template")
	}
	if lucy.EchoMaxScore["4"] != 86.518 || lucy.SubstatWeight["技能伤害加成"] != 0.9 || lucy.MainProps["1"]["攻击%"] != 0.6 {
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
	if chisa.EchoMaxScore["1"] != 74.977 || chisa.SubstatWeight["共鸣效率"] != 0.25 || chisa.SubstatWeight["技能伤害加成"] != 1.1 || chisa.SkillWeight[3] != 0.55 {
		t.Fatalf("unexpected 千咲 template: %+v", chisa)
	}

	sui, ok := resonatorTemplates["穗穗"]
	if !ok {
		t.Fatalf("expected 穗穗 score template")
	}
	if sui.EchoMaxScore["4"] != 51.991 || sui.MainProps["3"]["属性伤害加成"] != 0.28 || sui.SubstatWeight["生命%"] != 1.2 || sui.SubstatWeight["共鸣效率"] != 1 || sui.SubstatWeight["攻击"] != 0 {
		t.Fatalf("unexpected 穗穗 template: %+v", sui)
	}

	qingxiao, ok := resonatorTemplates["清宵"]
	if !ok {
		t.Fatalf("expected 清宵 score template")
	}
	if qingxiao.EchoMaxScore["4"] != 87.344 || qingxiao.MainProps["1"]["攻击%"] != 0.6 || qingxiao.SubstatWeight["暴击"] != 2 || qingxiao.SubstatWeight["技能伤害加成"] != 1.1 || qingxiao.SkillWeight[1] != 0.65 {
		t.Fatalf("unexpected 清宵 template: %+v", qingxiao)
	}
	if scoreEcho(EchoLog{}, "清霄", "4C", "暴击").Resonator != "清宵" {
		t.Fatalf("expected legacy 清霄 alias to resolve to 清宵 template")
	}

	score := scoreEcho(EchoLog{}, "未配置角色", "3C属伤", "")
	if score.Resonator != "通用" {
		t.Fatalf("expected unknown resonator to fall back to 通用, got %q", score.Resonator)
	}
}

func TestRebeccaFixedAttackSubstatScore(t *testing.T) {
	// 攻击 11.6%: substat number 2, value number 7.
	// 攻击固定值 60: substat number 5, value number 3.
	echo := EchoLog{
		Substat1: (1 << 2) | (1 << (7 + substatBitWidth)),
		Substat2: (1 << 5) | (1 << (3 + substatBitWidth)),
	}
	score := scoreEcho(echo, "丽贝卡", "4C", "")
	if score.Substat1 != 8.04 || score.Substat2 != 3.81 || score.SubstatAll != 11.85 {
		t.Fatalf("expected 丽贝卡 攻击 11.6%% + 攻击固定值 60 to score 8.04 + 3.81 = 11.85, got %.2f + %.2f = %.2f", score.Substat1, score.Substat2, score.SubstatAll)
	}
}

func TestMainstatScoresUseLevel25Values(t *testing.T) {
	tests := []struct {
		name          string
		resonator     string
		cost          string
		mainstat      string
		mainstatScore float64
		fixedScore    float64
		total         float64
	}{
		{name: "丽贝卡 1C 攻击%", resonator: "丽贝卡", cost: "1C", mainstat: "攻击%", mainstatScore: 6.44, fixedScore: 0, total: 6.44},
		{name: "丽贝卡 3C 导电伤害加成", resonator: "丽贝卡", cost: "3C属伤", mainstat: "导电伤害加成", mainstatScore: 5.32, fixedScore: 1.47, total: 6.79},
		{name: "丽贝卡 4C 暴击", resonator: "丽贝卡", cost: "4C", mainstat: "暴击", mainstatScore: 6.10, fixedScore: 1.73, total: 7.83},
		{name: "白芷 1C 生命%", resonator: "白芷", cost: "1C", mainstat: "生命%", mainstatScore: 9.01, fixedScore: 6.76, total: 15.77},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			score := scoreEcho(EchoLog{}, test.resonator, test.cost, test.mainstat)
			if score.MainstatScore != test.mainstatScore || score.MainstatScore2 != test.fixedScore || score.SubstatAll != test.total {
				t.Fatalf("expected main stats %.2f + %.2f = %.2f, got %.2f + %.2f = %.2f", test.mainstatScore, test.fixedScore, test.total, score.MainstatScore, score.MainstatScore2, score.SubstatAll)
			}
		})
	}
}

func TestXWUIDTemplateCountAndDuplicateNames(t *testing.T) {
	if len(xwuidTemplates) != 68 {
		t.Fatalf("expected 68 XW-UID templates, got %d", len(xwuidTemplates))
	}
	for _, name := range []string{"雷主", "洛瑟菈-霜渐", "洛瑟菈-声骸", "心", "锁暝"} {
		if _, ok := resonatorTemplates[name]; !ok {
			t.Fatalf("expected template %q", name)
		}
	}
}

func TestXWUIDSkillWeightAndPerEntryTruncation(t *testing.T) {
	template := scoreTemplateForResonator("达妮娅")
	encoded := int64(1<<12) | int64(1<<(3+substatBitWidth))
	weighted := substatValueScoreAt(2, encoded, "3C属伤", template) / template.EchoMaxScore["3"] * 50
	if got := truncateScore(weighted); got != 4.35 {
		t.Fatalf("expected XW-UID-truncated skill entry 4.35, got %.2f", got)
	}
}
