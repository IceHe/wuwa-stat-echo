# 高级统计与决策支持路线图

这份文档用于沉淀后续迭代方向，避免会话中断后丢失上下文。目标是把当前项目从“日志记录 + 基础统计”逐步推进到“可扩展统计平台 + 决策支持工具”。

当前重点关注 3 条线：

- 统计接口增量化，摆脱高频全表扫描
- 高级统计分析，支持置信区间、滚动窗口和显著性判断
- 用户进阶能力，支持养成决策和模拟器

## 总体策略

建议按以下顺序推进：

1. 先做统计底座，把统计口径和聚合层稳定下来
2. 再做高级统计，把“次数”升级为“有统计意义的结果”
3. 最后做决策支持和模拟器，把项目从分析工具升级为决策工具

原因：

- 当前多个接口仍依赖实时扫 `wuwa_tune_log` / `wuwa_echo_log`
- 如果底座不先抽象，后续高级分析和模拟器会反复返工
- 决策支持高度依赖稳定、可重建、可解释的统计层

## 一、统计接口增量化

### 目标

把以下接口从“请求时全表计算”逐步改造成“查询聚合结果”：

- `/tune_stats`
- `/counts/echo_dcrit`
- `/echo_logs/analysis`

### 设计原则

- `raw log` 是唯一真实源，聚合表都是可重建派生层
- 原始业务写入成功后，同步更新聚合表
- 必须支持全量重建和一致性校验
- 旧接口先兼容，内部实现再逐步切换

### 建议新增聚合表

#### 1. `agg_tune_substat_counts`

用途：

- 副词条分布
- 档位分布
- 孔位分布
- 用户维度统计
- 按天统计

建议字段：

```sql
bucket_type text not null,      -- all / day / rolling_snapshot
bucket_key text not null,       -- all / 2026-04-06 / snapshot id
user_id bigint not null,        -- 0 表示全站
substat int not null,
value int not null,             -- 0-7，额外允许 -1 表示 all
position int not null,          -- 0-4
count bigint not null default 0,
updated_at timestamptz not null default now(),
primary key (bucket_type, bucket_key, user_id, substat, value, position)
```

#### 2. `agg_echo_dcrit_counts`

用途：

- 双暴统计页
- 双暴档位组合分布

建议字段：

```sql
bucket_type text not null,
bucket_key text not null,
user_id bigint not null,
crit_rate_tier int not null,    -- -1 表示不存在
crit_dmg_tier int not null,     -- -1 表示不存在
count bigint not null default 0,
updated_at timestamptz not null default now(),
primary key (bucket_type, bucket_key, user_id, crit_rate_tier, crit_dmg_tier)
```

#### 3. `agg_echo_summary`

用途：

- 支撑 `/echo_logs/analysis`
- 资源消耗累计
- 目标词条命中率

建议字段：

```sql
bucket_type text not null,
bucket_key text not null,
user_id bigint not null,
target_bits int not null,
echo_count bigint not null default 0,
hit_count bigint not null default 0,
substat_total bigint not null default 0,
exp_total bigint not null default 0,
exp_recycled bigint not null default 0,
tuner_recycled bigint not null default 0,
updated_at timestamptz not null default now(),
primary key (bucket_type, bucket_key, user_id, target_bits)
```

#### 4. `agg_rebuild_jobs`

用途：

- 聚合重建任务
- 重建状态跟踪
- 统计层校验

建议字段：

```sql
id bigserial primary key,
job_type text not null,
status text not null,           -- pending/running/success/failed
started_at timestamptz,
finished_at timestamptz,
message text not null default ''
```

### 最近窗口的处理原则

`最近 100 / 500 / 1000 条` 这类窗口不是天然聚合维度，第一版不建议直接做成聚合表。

建议策略：

- 全量 / 按天统计：查聚合表
- 最近 N 条窗口：保留实时小范围计算
- 后续如果使用频率高，再补 snapshot 表

### 后端接口拆分建议

#### 原始业务接口

- `POST /echo_log`
- `PATCH /echo_log`
- `POST /echo_log/tune`
- `DELETE /echo_log/{id}`

职责：只负责写原始表。

#### 聚合维护接口

- `POST /admin/stats/rebuild`
- `GET /admin/stats/rebuild/{job_id}`
- `POST /admin/stats/reconcile`

职责：

- 全量重建聚合表
- 比对聚合表与原始表
- 修复统计漂移

#### 查询接口

建议新增统一统计接口，再让旧接口内部复用：

- `GET /stats/tune-distribution`
- `GET /stats/dcrit-distribution`
- `GET /stats/echo-summary`
- `GET /stats/variance-check`

旧接口如 `/tune_stats`、`/counts/echo_dcrit`、`/echo_logs/analysis` 先保留兼容。

### 后端代码结构建议

建议逐步拆成以下层次：

- `internal/goapp/repository`
  只负责原始表 / 聚合表读写
- `internal/goapp/stats`
  只负责统计口径和聚合更新
- `internal/goapp/analysis`
  只负责置信区间、显著性、趋势分析
- `internal/goapp/simulator`
  只负责蒙特卡洛模拟和结局分布

## 二、高级统计

### 目标

把“频数展示”升级为“有统计意义的分析”。

### 第一批能力

#### 1. 置信区间

为核心比例补充样本量和 95% 置信区间，例如：

- 副词条总占比
- 档位占比
- 双暴率
- 目标命中率

建议返回结构：

```json
{
  "count": 1180,
  "total": 4961,
  "rate": 0.2379,
  "ci95_low": 0.226,
  "ci95_high": 0.250
}
```

实现建议：

- 第一版使用 Wilson interval
- 比简单正态近似更稳，样本少时也更可靠

#### 2. 滚动窗口

建议统一支持以下窗口参数：

- `range=all`
- `range=last_100`
- `range=last_500`
- `range=last_1000`
- `range=day_7`
- `range=day_30`

这样前端可以统一切换统计视图，不需要理解底层实现差异。

#### 3. 显著性和偏差提示

建议先做用户可理解的结论层，而不是直接暴露 p-value。

建议返回：

```json
{
  "current_rate": 0.085,
  "baseline_rate": 0.112,
  "delta": -0.027,
  "sample_size": 200,
  "significance": "weak",
  "message": "近期双暴率偏低，但仍可能属于正常波动"
}
```

推荐分为 4 档：

- `none`
- `weak`
- `medium`
- `strong`

### 对比维度建议

第一版先支持：

- 个人最近窗口 vs 全站长期基线
- 当前窗口 vs 历史全量
- 某角色模板 vs 全站
- 某套装 vs 全站

### 前端页面建议

建议新增一个高级统计页，而不是继续堆叠在现有组件中。

页面模块建议：

- 全量 / 最近窗口切换
- 置信区间展示
- 显著性提示卡片
- 趋势图
- 个人 vs 全站对比

## 三、决策支持

### 目标

让系统回答“下一步该不该继续赌”。

### 第一版输入

```json
{
  "echo": { "...当前词条..." },
  "resonator": "弗洛洛",
  "cost": "4C",
  "goal": "毕业"
}
```

### 第一版输出

```json
{
  "current_score": 72.4,
  "percentile": 0.81,
  "effective_substat_count": 3,
  "locked_value": 0.67,
  "continue_to_next_prob": 0.41,
  "continue_to_finish_prob": 0.12,
  "expected_extra_tuner": 20,
  "expected_extra_exp": 16,
  "recommendation": "continue_once",
  "reasons": [
    "当前已具备双暴",
    "剩余孔位仍有 41% 概率命中目标词条",
    "同类历史分位已达 81%"
  ]
}
```

### `recommendation` 建议枚举

- `stop`
- `continue_once`
- `continue_to_end`
- `high_risk`

### 后端接口建议

- `POST /decision/echo-next-step`

### 前端页面建议

新增 `Decision Lab` 页面，展示：

- 当前评分
- 同类分位
- 当前有效词条数
- 继续一手 / 继续到底的收益和风险
- 建议卡片

## 四、模拟器

### 目标

支持对未来调谐路径做概率模拟，而不是只看历史统计。

### 第一版实现策略

第一版不追求数学闭式解，直接使用 Monte Carlo。

输入：

- 当前已有副词条
- 剩余孔位
- 目标角色模板
- 预算上限
- 模拟次数

输出：

- 达到小毕业 / 大毕业 / 满分阈值的概率
- 资源消耗分布
- 常见结局分布
- 继续赌 vs 立即止损对比

建议接口：

- `POST /simulator/echo-future`
- `POST /simulator/echo-compare`

示例输出：

```json
{
  "trials": 10000,
  "hit_prob": 0.128,
  "high_roll_prob": 0.036,
  "expected_score": 79.4,
  "expected_tuner_cost": 28,
  "expected_exp_cost": 22,
  "result_buckets": [
    {"label": "止步中等", "rate": 0.44},
    {"label": "小毕业", "rate": 0.39},
    {"label": "大毕业", "rate": 0.128},
    {"label": "神品", "rate": 0.006}
  ]
}
```

### 前端页面建议

新增 `Simulator` 页面，提供：

- 当前状态输入
- 目标模板选择
- 模拟参数输入
- 达标率和资源分布图
- 不同目标模板对比

## 五、推荐的 3 个 PR

### PR 1：统计底座

目标：

- 不改页面，先重构统计实现基础

内容：

- 新增聚合表 migration
- 新增全量重建脚本
- 抽离 `stats service`
- 旧接口内部切到新 service
- 增加统计口径回归测试

验收：

- 旧接口返回保持兼容
- 聚合结果与旧逻辑一致
- 查询性能优于全表扫描

### PR 2：高级统计

目标：

- 从“看次数”升级为“看意义”

内容：

- 新增 `/stats/tune-distribution`
- 新增 `/stats/variance-check`
- 返回置信区间、窗口对比、显著性提示
- 前端新增高级统计页面

验收：

- 页面支持全量 / 最近窗口切换
- 核心比例带样本量和置信区间
- 能看个人 vs 全站偏差

### PR 3：决策支持 + 模拟器 v1

目标：

- 做出一个真正有产品感的进阶能力

内容：

- 新增 `/decision/echo-next-step`
- 新增 `/simulator/echo-future`
- 新增 Decision Lab 页面
- 新增 Simulator 页面

验收：

- 用户能看到继续 / 止损建议
- 能对未来路径做概率模拟

## 六、推荐执行顺序

建议按以下顺序落地：

1. migration 和 rebuild 脚本
2. `stats service` 抽象
3. 统计回归测试
4. 高级统计新接口
5. 决策支持
6. 模拟器

## 七、最小可落地版本

如果只做第一步，建议先收敛到以下最小范围：

- 先只做 `agg_tune_substat_counts`
- 先只重构 `/tune_stats`
- 顺手补：
  - 95% 置信区间
  - 最近 100 / 500 窗口
  - 个人 vs 全站对比

这是当前收益最高、实现复杂度最低的一步。
