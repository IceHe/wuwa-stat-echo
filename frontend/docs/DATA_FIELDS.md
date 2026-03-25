# 前端数据字段文档

## 目标

说明前端页面中最常见的数据对象和字段，帮助后续维护者理解组件状态与后端返回结构。

## 1. `echoLog`

主要出现在：

- `Echo.vue`
- `EchoBoard.vue`
- `FindEcho.vue`

字段口径基本继承后端 `EchoLog`：

| 字段 | 说明 |
|---|---|
| `id` | 当前声骸 ID |
| `clazz` | 套装名称 |
| `user_id` | 玩家 ID |
| `pos` | 当前正在操作的孔位，前端扩展字段 |
| `substat1..5` | 各孔位副词条编码值 |
| `substat_all` | 所有副词条类型位图 |
| `s1_desc..s5_desc` | 各孔位展示文案 |
| `pos_total` | 当前孔位剩余样本数，来自后端扩展返回 |

## 2. `template`

主要出现在 `Echo.vue`，用于分析筛选：

| 字段 | 说明 |
|---|---|
| `clazz` | 当前套装 |
| `user_id` | 当前玩家 |
| `after_echo_id` | 统计起始边界 |
| `before_echo_id` | 统计结束边界 |

## 3. `scoreTemplate`

用于评分分析：

| 字段 | 说明 |
|---|---|
| `resonator` | 共鸣者模板名称 |
| `cost` | Cost 主词条类型，如 `1C`、`3C`、`4C` |

## 4. `recentTuneStats`

来自 `/tune_stats`，常见字段：

| 字段 | 说明 |
|---|---|
| `data_total` | 统计样本总量 |
| `substat_dict` | 各副词条统计对象 |
| `substat_distance` | 距上次出现的间隔 |
| `substat_pos_total` | 孔位分布矩阵 |
| `position_total` | 各孔位总数 |

## 5. `echoAnalysis`

来自 `/analyze_echo`，不同组件消费的重点不同，通常包括：

| 字段 | 说明 |
|---|---|
| `score` | 声骸评分对象 |
| `substat_dict` | 当前孔位候选副词条统计 |
| `two_crit_percent` | 双暴相关概率展示 |

### `score`

| 字段 | 说明 |
|---|---|
| `substat1..5` | 每个已出副词条的评分 |
| `substat_all` | 总评分 |

## 6. `currentUser` / `allUsers`

来自 `/echo_logs/analysis`：

| 字段 | 说明 |
|---|---|
| `target` | 命中目标词条的声骸数 |
| `target_echo_distance` | 距离上次命中目标隔了多少个声骸 |
| `target_substat_distance` | 距离上次命中目标隔了多少个副词条 |
| `target_avg_echo` | 平均多少个声骸命中一次 |
| `target_avg_substat` | 平均多少个副词条命中一次 |
| `tuner_consumed` | 总调谐器消耗估算 |
| `tuner_consumed_avg` | 平均调谐器消耗 |
| `exp_consumed` | 总经验消耗估算 |
| `exp_consumed_avg` | 平均经验消耗 |

## 7. `SUBSTAT`

来自 `src/stores/constants.ts` 的副词条字典数组。

每项一般包含：

| 字段 | 说明 |
|---|---|
| `num` | 副词条编号 |
| `name` | 中文名 |
| `font_color` | 页面展示颜色 |
| `bitmap` | 类型位图 |

## 8. `SUBSTAT_VALUE_MAP`

键为副词条编号，值为档位数组。

每个档位元素一般包含：

| 字段 | 说明 |
|---|---|
| `substat_number` | 副词条编号 |
| `value_number` | 档位编号 |
| `desc` | 展示简写 |
| `desc_full` | 完整展示文案 |

## 9. URL 参数

前端部分状态会同步进路由查询参数：

| 参数 | 来源组件 | 说明 |
|---|---|---|
| `echo_id` | `Echo.vue` | 当前声骸 ID |
| `user_id` | `Echo.vue` | 当前玩家 ID |
| `clazz` | `Echo.vue` | 当前套装 |
| `resonator` | `Echo.vue` | 当前评分模板 |
| `cost` | `Echo.vue` | 当前 Cost |
| `after_echo_id` | `Echo.vue` | 分析下界 |
| `before_echo_id` | `Echo.vue` | 分析上界 |
