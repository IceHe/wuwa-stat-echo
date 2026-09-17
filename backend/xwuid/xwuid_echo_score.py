"""Standalone XutheringWavesUID-compatible legacy echo scoring.

Data and behavioral sources
---------------------------
* Template data is imported from :mod:`xwuid_echo_templates`.  That file is a
  snapshot of the public XutheringWavesUID-Resources ``calc*.json`` files and
  records the exact upstream repository and commit.
* The normalization formula is printed by XutheringWavesUID in
  ``wutheringwaves_charinfo/draw_char_card.py``:
  ``value * weight / unaligned_max_score * 50``.
* Function signatures, the two-main-stat boundary, skill-weight mapping and
  two-decimal truncation were verified by black-box comparison against the
  publicly released CPython 3.10 ``waves_build.calculate`` extension from the
  same resource snapshot.

This is an independent compatibility implementation.  It does not import,
bundle, decompile, or require XW-UID's compiled extension.  Copy this file and
``xwuid_echo_templates.py`` together to reuse it in another Python project.
"""

from __future__ import annotations

from copy import deepcopy
from dataclasses import dataclass
import math
import re
from typing import Any, Iterable, Mapping, Sequence

try:  # Works both as two loose files and as files inside a Python package.
    from .xwuid_echo_templates import CONDITIONS, TEMPLATES
except ImportError:  # pragma: no cover - exercised when copied as loose files
    from xwuid_echo_templates import CONDITIONS, TEMPLATES


SCORE_PER_ECHO = 50.0
SCORE_PER_LOADOUT = 250.0
GRADE_NAMES = ("c", "b", "a", "s", "ss", "sss")
COST_INDEX = {1: 0, 3: 1, 4: 2}
ELEMENT_NAMES = {"冷凝", "热熔", "导电", "气动", "衍射", "湮灭"}
SKILL_WEIGHT_INDEX = {
    "普攻伤害加成": 0,
    "重击伤害加成": 1,
    "共鸣技能伤害加成": 2,
    "共鸣解放伤害加成": 3,
}


@dataclass(frozen=True)
class Prop:
    """Minimal replacement for XW-UID's ``Props`` model."""

    attributeName: str
    attributeValue: str | int | float


@dataclass(frozen=True)
class Echo:
    """One echo for :func:`calc_loadout_score`."""

    cost: int
    props: Sequence[Prop | Mapping[str, Any] | Sequence[Any]]


@dataclass(frozen=True)
class ScoreResult:
    score: float
    grade: str


def _variant_from_filename(filename: str) -> str:
    stem = filename.removesuffix(".json")
    return "default" if stem == "calc" else stem.removeprefix("calc-")


def _compare(actual: Any, operator: str, expected: Any) -> bool:
    if operator == "=":
        return actual == expected
    if operator == "!=":
        return actual != expected
    if operator == "in":
        return any(item in (actual or []) for item in expected) if isinstance(actual, (list, tuple, set)) else actual in expected
    if operator == "!in":
        return not _compare(actual, "in", expected)
    if operator == "<":
        return actual < expected
    if operator == ">":
        return actual > expected
    if operator == "<=":
        return actual <= expected
    if operator == ">=":
        return actual >= expected
    raise ValueError(f"unsupported template-condition operator: {operator!r}")


def _condition_matches(context: Mapping[str, Any], expression: Mapping[str, Any]) -> bool:
    operator = expression.get("op")
    children = expression.get("sub")
    if children is not None:
        if operator == "&&":
            return all(_condition_matches(context, child) for child in children)
        if operator == "||":
            return any(_condition_matches(context, child) for child in children)
        raise ValueError(f"unsupported compound operator: {operator!r}")
    return _compare(context.get(expression.get("key")), str(operator), expression.get("value"))


def list_templates() -> dict[str, tuple[str, ...]]:
    """Return all embedded character IDs and their available variants."""
    return {char_id: tuple(variants) for char_id, variants in TEMPLATES.items()}


def get_calc_map(
    ctx: Mapping[str, Any] | None,
    char_name: str,
    char_id: str | int,
    modal: str = "",
    *,
    variant: str | None = None,
) -> dict[str, Any]:
    """Select and return a character scoring template.

    The first four parameters mirror XW-UID's public ``get_calc_map`` call.
    ``char_name`` is accepted for API compatibility; template selection is
    keyed by character ID.  Unknown IDs fall back to the upstream ``default``
    template.
    """
    char_key = str(char_id)
    variants = TEMPLATES.get(char_key, TEMPLATES["default"])
    context = dict(ctx or {})
    if modal:
        context["modal"] = modal

    selected = variant
    if selected is None:
        for expression in CONDITIONS.get(char_key, ()):
            if _condition_matches(context, expression):
                selected = _variant_from_filename(expression["choose"])
                break
    if selected is None:
        selected = modal if modal in variants else "default"
    template = variants.get(selected) or variants.get("default") or TEMPLATES["default"]["default"]
    return deepcopy(template)


def as_number(value: str | int | float) -> float:
    """Parse API values such as ``'10.5%'``, ``'1,200'`` and plain numbers."""
    if isinstance(value, bool):
        raise TypeError("boolean is not a scoreable attribute value")
    if isinstance(value, (int, float)):
        return float(value)
    match = re.search(r"[-+]?\d+(?:\.\d+)?", value.replace(",", ""))
    if not match:
        raise ValueError(f"unsupported attribute value: {value!r}")
    return float(match.group())


def _read_prop(prop: Prop | Mapping[str, Any] | Sequence[Any] | Any) -> tuple[str, str | int | float]:
    if isinstance(prop, Mapping):
        name = prop.get("attributeName", prop.get("name"))
        value = prop.get("attributeValue", prop.get("value"))
    elif isinstance(prop, Sequence) and not isinstance(prop, (str, bytes)) and len(prop) == 2:
        name, value = prop
    else:
        name = getattr(prop, "attributeName", getattr(prop, "name", None))
        value = getattr(prop, "attributeValue", getattr(prop, "value", None))
    if not isinstance(name, str) or value is None:
        raise TypeError("property must expose attributeName/attributeValue, name/value, or be a 2-item sequence")
    return name, value


def _main_name(name: str) -> str:
    if name.endswith("伤害加成") and name.removesuffix("伤害加成") in ELEMENT_NAMES:
        return "属性伤害加成"
    return name


def _entry_weight(index: int, name: str, cost: int, calc_map: Mapping[str, Any]) -> float:
    # XW-UID's API emits two main properties first: the selectable main stat
    # and its fixed companion stat.  Rolled sub-stats start at index 2.
    if index < 2:
        return float(calc_map.get("main_props", {}).get(str(cost), {}).get(_main_name(name), 0.0))

    sub_weights = calc_map.get("sub_props", {})
    if name in SKILL_WEIGHT_INDEX:
        generic_weight = float(sub_weights.get("技能伤害加成", 0.0))
        skill_weights = calc_map.get("skill_weight", [0.0, 0.0, 0.0, 0.0])
        return generic_weight * float(skill_weights[SKILL_WEIGHT_INDEX[name]])
    return float(sub_weights.get(name, 0.0))


def get_max_score(cost: int, calc_map: Mapping[str, Any]) -> float:
    """Return the template's unaligned maximum raw score for an echo cost."""
    try:
        return float(calc_map["score_max"][COST_INDEX[int(cost)]])
    except (KeyError, IndexError, TypeError) as exc:
        raise ValueError(f"unsupported echo cost or malformed template: {cost!r}") from exc


def _truncate_2(value: float) -> float:
    """Match XW-UID's per-entry two-decimal truncation (not rounding)."""
    return math.trunc((value + 1e-12) * 100.0) / 100.0


def calc_phantom_entry(
    index: int,
    prop: Prop | Mapping[str, Any] | Sequence[Any] | Any,
    cost: int,
    calc_map: Mapping[str, Any],
    char_attr: str = "",
) -> tuple[float, float]:
    """Return ``(raw_weighted_value, normalized_points)`` for one property."""
    del char_attr  # Elemental names are recognized directly; retained for API compatibility.
    name, formatted_value = _read_prop(prop)
    raw = as_number(formatted_value) * _entry_weight(index, name, int(cost), calc_map)
    maximum = get_max_score(int(cost), calc_map)
    points = _truncate_2(raw / maximum * SCORE_PER_ECHO) if maximum else 0.0
    return raw, points


def _grade(score: float, fractions: Sequence[float], maximum: float) -> str:
    for fraction, name in reversed(tuple(zip(fractions, GRADE_NAMES))):
        if score >= float(fraction) * maximum:
            return name
    return "c"


def get_echo_grade(score: float, cost: int, calc_map: Mapping[str, Any]) -> str:
    fractions = calc_map.get("props_grade", [])[COST_INDEX[int(cost)]]
    return _grade(float(score), fractions, SCORE_PER_ECHO)


def calc_phantom_score(
    char_id: str | int,
    prop_list: Iterable[Prop | Mapping[str, Any] | Sequence[Any] | Any],
    cost: int,
    calc_map: Mapping[str, Any] | None,
) -> tuple[float, str]:
    """Score one echo and return ``(points, grade)`` like XW-UID."""
    del char_id  # Kept to match XW-UID's public signature.
    if not calc_map:
        return 0.0, "c"
    score = round(sum(calc_phantom_entry(i, prop, cost, calc_map)[1] for i, prop in enumerate(prop_list)), 2)
    return score, get_echo_grade(score, cost, calc_map)


def get_total_score_bg(char_name: str, score: float, calc_map: Mapping[str, Any] | None) -> str:
    """Return the C/B/A/S/SS/SSS grade for a five-echo, 250-point total."""
    del char_name  # Kept to match XW-UID's public signature.
    if not calc_map:
        return "c"
    fractions = calc_map.get("total_grade") or [0.0, 0.48, 0.60, 0.70, 0.78, 0.84]
    return _grade(float(score), fractions, SCORE_PER_LOADOUT)


def calc_loadout_score(
    char_id: str | int,
    echoes: Iterable[Echo | Mapping[str, Any]],
    calc_map: Mapping[str, Any],
) -> ScoreResult:
    """Score an arbitrary collection of echoes and grade it on the 250 scale."""
    total = 0.0
    for echo in echoes:
        if isinstance(echo, Mapping):
            cost, props = int(echo["cost"]), echo["props"]
        else:
            cost, props = echo.cost, echo.props
        total += calc_phantom_score(char_id, props, cost, calc_map)[0]
    total = round(total, 2)
    return ScoreResult(total, get_total_score_bg("", total, calc_map))


__all__ = [
    "CONDITIONS",
    "Echo",
    "Prop",
    "ScoreResult",
    "TEMPLATES",
    "as_number",
    "calc_loadout_score",
    "calc_phantom_entry",
    "calc_phantom_score",
    "get_calc_map",
    "get_echo_grade",
    "get_max_score",
    "get_total_score_bg",
    "list_templates",
]
