#!/usr/bin/env python3
"""Generate the Go XW-UID template snapshot from the Python source data."""

from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from xwuid.xwuid_echo_templates import SOURCE_COMMIT, TEMPLATES


def go(value):
    if isinstance(value, str):
        return json.dumps(value, ensure_ascii=False)
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, (int, float)):
        return str(value)
    if isinstance(value, list):
        if value and isinstance(value[0], list):
            return "[][]float64{" + ", ".join(go(item) for item in value) + "}"
        return "[]float64{" + ", ".join(go(item) for item in value) + "}"
    if isinstance(value, dict):
        pairs = []
        for key, item in value.items():
            pairs.append(f"{go(key)}: {go(item)}")
        value_type = "map[string]float64" if all(isinstance(item, (int, float)) for item in value.values()) else "map[string]map[string]float64"
        return f"{value_type}{{" + ", ".join(pairs) + "}"
    raise TypeError(type(value))


lines = [
    "package goapp",
    "",
    f"// Generated from xwuid.xwuid_echo_templates; source commit {SOURCE_COMMIT}.",
    "",
    "type xwuidTemplate struct {",
    '\tID string `json:"id"`',
    '\tVariant string `json:"variant"`',
    '\tName string `json:"name"`',
    '\tSourceName string `json:"source_name"`',
    '\tScoreMax []float64 `json:"score_max"`',
    '\tPropsGrade [][]float64 `json:"props_grade"`',
    '\tTotalGrade []float64 `json:"total_grade"`',
    '\tMainProps map[string]map[string]float64 `json:"main_props"`',
    '\tSubProps map[string]float64 `json:"sub_props"`',
    '\tSkillWeight []float64 `json:"skill_weight"`',
    "}",
    "",
    "var xwuidTemplates = []xwuidTemplate{",
]

for role_id, variants in TEMPLATES.items():
    for variant, data in variants.items():
        source_name = data["name"]
        display_name = source_name.removesuffix("-通用")
        lines.extend([
            "\t{",
            f'\t\tID: {go(role_id)}, Variant: {go(variant)}, Name: {go(display_name)}, SourceName: {go(source_name)},',
            f'\t\tScoreMax: {go(data["score_max"])},',
            f'\t\tPropsGrade: {go(data["props_grade"])},',
            f'\t\tTotalGrade: {go(data["total_grade"])},',
            f'\t\tMainProps: {go(data["main_props"])},',
            f'\t\tSubProps: {go(data["sub_props"])},',
            f'\t\tSkillWeight: {go(data["skill_weight"])},',
            "\t},",
        ])

lines.extend([
    "}",
    "",
    "var xwuidTemplateByKey = func() map[string]xwuidTemplate {",
    "\tresult := make(map[string]xwuidTemplate, len(xwuidTemplates))",
    "\tfor _, template := range xwuidTemplates {",
    "\t\tresult[template.Name] = template",
    "\t}",
    "\treturn result",
    "}()",
    "",
])

(ROOT / "internal/goapp/xwuid_templates_generated.go").write_text("\n".join(lines))
