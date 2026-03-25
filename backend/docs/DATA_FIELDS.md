# 数据字段文档

## 目标

本文件说明后端最核心的数据字段语义，方便接口联调、数据库排查和前端渲染统一口径。

## 1. `EchoLog`

定义位置：[model.py](../model.py)

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `int` | 声骸记录主键 |
| `substat1` ~ `substat5` | `int` | 第 1~5 个副词条编码值 |
| `substat_all` | `int` | 所有副词条类型汇总位图 |
| `s1_desc` ~ `s5_desc` | `str` | 对应副词条展示文本 |
| `clazz` | `str` | 套装名称 |
| `user_id` | `int` | 当前玩家 ID |
| `deleted` | `int` | 软删除标记 |
| `tuned_at` | `datetime?` | 调谐时间 |
| `created_at` | `datetime?` | 创建时间 |
| `updated_at` | `datetime?` | 更新时间 |

### 运行时扩展字段

以下字段不在数据库表中，但部分接口会在返回里追加：

| 字段 | 来源 | 说明 |
|---|---|---|
| `pos_total` | `GET /echo_log/{id}` | 当前孔位剩余样本总量 |

## 2. `SubstatLog`

定义位置：[model.py](../model.py)

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `int` | 调谐日志主键 |
| `substat` | `int` | 副词条类型编号 |
| `value` | `int` | 档位编号 |
| `position` | `int` | 孔位编号，0 表示第 1 孔 |
| `echo_id` | `int` | 所属声骸记录 ID |
| `user_id` | `int` | 玩家 ID |
| `timestamp` | `datetime?` | 调谐记录时间 |
| `deleted` | `int?` | 软删除标记 |

## 3. `tune_stats` 统计结构

主要由：

- `GET /tune_stats`
- `shared.tune_stats`

共同产出。

顶层字段：

| 字段 | 说明 |
|---|---|
| `data_total` | 当前统计样本总数 |
| `substat_dict` | 按副词条编号组织的统计对象 |
| `substat_distance` | 每类副词条距离最近一次出现的样本间隔 |
| `substat_pos_total` | 每类副词条在每个孔位的次数矩阵 |
| `position_total` | 每个孔位的样本总数 |

### `substat_dict[x]`

| 字段 | 说明 |
|---|---|
| `number` | 副词条编号 |
| `name` | 英文名 |
| `name_cn` | 中文名 |
| `total` | 出现总次数 |
| `percent` | 占全部样本比例 |
| `value_dict` | 各档位统计对象 |
| `cur_pos_percent` | 当前孔位场景下的展示值，主要供前端使用 |

### `value_dict[y]`

| 字段 | 说明 |
|---|---|
| `value_number` | 档位编号 |
| `value_desc` | 档位简写，如 `8.1%` |
| `value_desc_full` | 档位完整文案，如 `暴击 8.1%` |
| `total` | 该档位出现次数 |
| `percent` | 该档位占该副词条全部样本的比例 |
| `percent_substat` | 当前副词条内部的档位分布比例 |
| `position_dict` | 按孔位拆分的统计对象 |

## 4. 资源分析字段

`GET /echo_logs/analysis` 返回的数据包含：

| 字段 | 说明 |
|---|---|
| `target` | 命中目标词条组合的声骸数 |
| `target_echo_distance` | 距离上次出货隔了多少个声骸 |
| `target_substat_distance` | 距离上次出货隔了多少个副词条 |
| `target_avg_echo` | 平均多少个声骸出一次 |
| `target_avg_substat` | 平均多少个副词条出一次 |
| `tuner_consumed` | 总调谐器消耗估算 |
| `tuner_consumed_avg` | 平均调谐器消耗 |
| `exp_consumed` | 总经验消耗估算 |
| `exp_consumed_avg` | 平均经验消耗 |

## 5. 评分分析字段

`POST /analyze_echo` 的返回中，常用字段包括：

| 字段 | 说明 |
|---|---|
| `score.substat1` ~ `score.substat5` | 各副词条评分 |
| `score.substat_all` | 总评分 |
| `two_crit_percent` | 双暴概率或双暴命中率展示值 |
| `substat_dict` | 当前孔位下各候选副词条的概率统计 |

## 6. 前后端协作注意点

- `position` 在数据库中是 `0-4`，但页面展示时常被理解成第 `1-5` 孔
- `substat1..5` 是完整编码值，不是简单编号
- `substat_all` 只表示副词条类型集合，不携带档位信息
- `s1_desc..s5_desc` 是展示字段，不能替代编码字段参与分析
