# XW-UID 声骸评分兼容模块

这是一组可独立复用的、纯 Python 的旧版声骸评分模块。目录中只有两份 Python 文件：

- `xwuid_echo_templates.py`：角色/模态评分模板、主副词条权重、分数上限、评级阈值与条件选模板规则。
- `xwuid_echo_score.py`：模板选择、属性解析、单词条/单声骸/五件套评分及评级逻辑。

不依赖 XW-UID 主项目、`gsuid_core` 或编译版 `.pyd`；将这两个文件一起复制到其他 Python 项目即可使用。

## 数据来源与范围

模板来自公开资源仓库：

```text
仓库: https://cnb.cool/loping151/XutheringWavesUID-Resources
提交: 2e8a2510c195a42f1fb4090a3bb2768bb85b4543
路径: XutheringWavesUID/resource/map/character/<角色 ID>/calc*.json
快照: 2026-09-14
```

当前快照内嵌 66 份角色/模态模板，也包含洛瑟莉根据套装与模态选择模板的规则。

本模块复刻的是 XW-UID 的 **旧声骸评分**：每件满分 50，五件满分 250。它不是新版“综合评分”（150 分）或伤害模拟器。

计分规则为：

```text
原始值 = 词条数值 × 当前模板词条权重
词条分 = truncate(原始值 / 对应 Cost 的 score_max × 50, 2)
单件分 = 所有词条分之和
总分 = 所有声骸单件分之和
```

实现细节与公开版行为保持一致：前两个属性是主词条（可选主词条与固定附加属性），从第三项开始才是副词条；普攻、重击、共鸣技能、共鸣解放副词条会分别乘模板的 `skill_weight`；单词条分数保留两位时采用截断而非四舍五入。

## 导入

运行入口的 `PYTHONPATH` 包含 `backend` 时：

```python
from xwuid.xwuid_echo_score import (
    Echo,
    Prop,
    calc_loadout_score,
    calc_phantom_score,
    get_calc_map,
)
```

若把两个 `.py` 文件直接复制到同一目录，也可直接：

```python
from xwuid_echo_score import Prop, calc_phantom_score, get_calc_map
```

## 单件声骸评分示例

`props` 的顺序很重要：前两项必须是主词条，之后才是副词条。

```python
from xwuid.xwuid_echo_score import Prop, calc_phantom_score, get_calc_map

# 1208 = 嘉贝莉娜；ctx 是用于选择特殊模板的上下文，普通角色可传 {}。
template = get_calc_map(ctx={}, char_name="嘉贝莉娜", char_id="1208")

props = [
    Prop("暴击", "22%"),       # 主词条
    Prop("攻击", "150"),       # 固定附加主词条
    Prop("暴击", "10.5%"),     # 以下均为副词条
    Prop("暴击伤害", "21%"),
    Prop("攻击%", "11.6%"),
    Prop("重击伤害加成", "11.6%"),
    Prop("攻击", "60"),
]

score, grade = calc_phantom_score(
    char_id="1208",
    prop_list=props,
    cost=4,
    calc_map=template,
)
print(score, grade)  # 49.96, 'sss'
```

词条也可以是字典或二元组：

```python
{"attributeName": "暴击", "attributeValue": "10.5%"}
("暴击", "10.5%")
```

## 五件声骸总评分示例

```python
from xwuid.xwuid_echo_score import Echo, Prop, calc_loadout_score, get_calc_map

template = get_calc_map({}, "嘉贝莉娜", "1208")
echoes = [
    Echo(cost=4, props=[Prop("暴击", "22%"), Prop("攻击", "150"), ...]),
    Echo(cost=3, props=[...]),
    Echo(cost=3, props=[...]),
    Echo(cost=1, props=[...]),
    Echo(cost=1, props=[...]),
]

result = calc_loadout_score("1208", echoes, template)
print(result.score, result.grade)  # 0～250，C～SSS
```

## 特殊模态模板示例

`get_calc_map` 会按内嵌条件自动选择模板；也可以传入 `modal` 或 `variant` 显式指定。

```python
# 洛瑟莉：羽落空尘之歌 + 声骸模态 → calc-phantom-featherfall.json
template = get_calc_map(
    ctx={"sonata_5": ["羽落空尘之歌"]},
    char_name="洛瑟莉",
    char_id="1109",
    modal="phantom",
)

# 直接选择同一角色的某个模板变体。
template = get_calc_map({}, "洛瑟莉", "1109", variant="phantom-featherfall")
```

## 更新模板

当上游评分模板更新时，从上述仓库同一路径重新收集所有 `calc*.json` 与 `condition.json`，再生成 `xwuid_echo_templates.py` 中的 `TEMPLATES`、`CONDITIONS` 即可。不要手工修改单个角色权重，否则文件头的提交号将不再能准确描述数据来源。

评分逻辑本身通常无需随角色更新；角色新增或权重变动应只修改模板数据文件。

## 给后续 AI 的更新操作指南

更新前先确认本目录当前 `xwuid_echo_templates.py` 的 `SOURCE_COMMIT`，并将新旧提交号、检查日期写回文件头。不要仅凭角色名称或其他第三方评分表补模板。

### 1. 获取模板和权重变更

拉取以下公开资源仓库的 `main` 分支：

```text
仓库: https://cnb.cool/loping151/XutheringWavesUID-Resources.git
文件: XutheringWavesUID/resource/map/character/**/calc*.json
文件: XutheringWavesUID/resource/map/character/**/condition.json
```

其中 `calc*.json` 是每名角色及特殊模态的主词条权重、副词条权重、`skill_weight`、`score_max`、单件/总分评级阈值；`condition.json` 决定按套装、模态等上下文选用哪个 `calc-*.json`。应比较新旧提交在这些路径下的 diff，然后 **重新汇总全部文件** 到 `xwuid_echo_templates.py` 的 `TEMPLATES` 与 `CONDITIONS`，而不是只粘贴本次新增角色。

生成后至少校验：

```python
from xwuid.xwuid_echo_templates import TEMPLATES
assert all(len(data["score_max"]) == 3 for variants in TEMPLATES.values() for data in variants.values())
assert all(len(data["props_grade"]) == 3 for variants in TEMPLATES.values() for data in variants.values())
assert all(len(data["skill_weight"]) == 4 for variants in TEMPLATES.values() for data in variants.values())
```

### 2. 判断评分逻辑是否变更

同时检查主插件仓库的 `main` 分支：

```text
仓库: https://github.com/Loping151/XutheringWavesUID.git
文件: XutheringWavesUID/utils/calculate.py
文件: XutheringWavesUID/wutheringwaves_charinfo/draw_char_card.py
文件: XutheringWavesUID/utils/expression_evaluator.py
```

重点查看 `calc_phantom_entry`、`calc_phantom_score`、`get_calc_map` 的调用参数、展示公式文本、主/副词条顺序与条件表达式语义。实际运行实现发布在资源仓库的：

```text
XutheringWavesUID/resource/build/<平台>/waves_build/calculate.<平台扩展名>
```

该文件通常是编译扩展而非 `.py` 源码。若公开调用层或二进制行为发生变化，更新 `xwuid_echo_score.py`，并以匹配 Python/平台的扩展做黑盒对拍；至少覆盖 Cost 1/3/4、两个主词条、暴击/暴伤/百分比/固定值副词条、四类技能伤害副词条、单件评级和五件总评级。

### 3. 本项目中允许改动的目标

模板数据变动应只改：

```text
backend/xwuid/xwuid_echo_templates.py
```

公式或兼容接口变动应只改：

```text
backend/xwuid/xwuid_echo_score.py
```

最后更新本 README 的上游提交号与快照日期，执行：

```text
python -m py_compile backend/xwuid/xwuid_echo_templates.py backend/xwuid/xwuid_echo_score.py
```

并运行本 README 的单件、五件和特殊模态示例后再提交。
