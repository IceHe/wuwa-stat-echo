# 前端接口文档

## 概览

前端不提供独立后端接口，而是消费后端 HTTP 服务。本文档记录当前页面实际依赖的 HTTP 接口与 WebSocket。

默认后端地址：

```text
http://${API_SERV}
ws://${API_SERV}/ws
```

## HTTP 接口

### 声骸记录

| 方法 | 路径 | 主要调用组件 | 作用 |
|---|---|---|---|
| `GET` | `/echo_logs?page_size=` | `EchoLogs` `EchoLogsWs` | 获取声骸列表 |
| `GET` | `/echo_log/{id}` | `Echo` `EchoBoard` | 获取单个声骸或最新声骸 |
| `POST` | `/echo_log` | `Echo` | 新建声骸 |
| `PATCH` | `/echo_log` | `Echo` | 更新声骸 |
| `DELETE` | `/echo_log/{id}` | `EchoLogRow` | 软删除声骸 |
| `POST` | `/echo_log/{id}/recover` | `EchoLogRow` | 恢复声骸 |
| `POST` | `/echo_log/find` | `FindEcho` | 按孔位词条或副词条集合查找声骸 |
| `POST` | `/echo_log/recent_search` | `RecentEchoesView` | 最近声骸分页搜索 |
| `DELETE` | `/echo_log/{echoId}/substat_pos/{pos}` | `Echo` | 删除声骸某一孔位对应记录 |
| `GET` | `/echo_logs/analysis` | `Echo` `EchoBoard` | 目标词条、资源和间隔分析 |

### 副词条调谐记录

| 方法 | 路径 | 主要调用组件 | 作用 |
|---|---|---|---|
| `GET` | `/substat_logs?page_size=` | `SubstatLogs` | 获取调谐日志列表 |
| `POST` | `/tune_log` | `Substat` `Echo` | 新增单条调谐记录 |
| `POST` | `/tune_log/{id}/delete` | `SubstatLogRow` | 硬删除调谐记录 |

### 统计与分析

| 方法 | 路径 | 主要调用组件 | 作用 |
|---|---|---|---|
| `GET` | `/tune_stats` | `Substat` `SubstatAnalysis` `Echo` `EchoBoard` | 获取副词条统计 |
| `POST` | `/analyze_echo` | `Echo` | 计算当前声骸评分与候选概率 |
| `GET` | `/counts/echo_dcrit` | `EchoDcritCount` | 双暴档位统计 |

## WebSocket

| 路径 | 主要调用组件 | 用途 |
|---|---|---|
| `/ws` | `SubstatAnalysis` `EchoLogsWs` `EchoBoard` | 接收后端广播后刷新列表或统计 |

## 常见查询参数

### `/tune_stats`

- `size`
- `user_id`
- `after_id`
- `before_id`

### `/echo_logs/analysis`

- `size`
- `user_id`
- `target_bits`
- `after_echo_id`
- `before_echo_id`

### `/counts/echo_dcrit`

- `after_id`
- `before_id`

## 搜索接口约定

### `/echo_log/find` 与 `/echo_log/recent_search`

两个接口共用一套搜索模式字段：

- `search_mode = positional`
- `search_mode = substat_set`

`positional` 模式：

- 使用 `substat1..5`
- 按指定孔位上的词条与档位搜索

`substat_set` 模式：

- 使用 `substat_all_mask`
- 只按副词条类型集合搜索
- 忽略孔位
- 忽略档位

副词条集合搜索当前采用“包含所选”语义：

- 例如 `substat_all_mask` 选择“暴击 + 暴伤”时，返回所有包含双暴的声骸
- 不要求这两个词条出现在固定孔位
- 不要求这两个词条是固定档位

前端页面行为与接口保持一致：

- `RecentEchoesView` 中，`user_id = 0` 表示浏览全局最近声骸，`user_id > 0` 表示按玩家筛选
- `operator_id = 0` 表示不按操作者筛选，`operator_id > 0` 表示仅看该操作者录入的声骸，仅管理员可用
- 选择“孔位搜索”时，只发送 `substat1..5`
- 选择“副词条搜索”时，只发送 `substat_all_mask`
- 两种模式互斥，不会同时生效

## 页面与接口映射

### `EchoView`

依赖：

- `/echo_log`
- `/echo_logs`
- `/echo_logs/analysis`
- `/tune_log`
- `/tune_stats`
- `/analyze_echo`
- `/ws`

### `SubstatView`

依赖：

- `/tune_log`
- `/substat_logs`
- `/tune_stats`
- `/ws`

### `EchoBoardView`

依赖：

- `/echo_log/{id}`
- `/echo_logs/analysis`
- `/tune_stats`
- `/ws`

### `RecentEchoesView`

依赖：

- `/echo_log/recent_search`

### `EchoDcritCountView`

依赖：

- `/counts/echo_dcrit`

## 当前前端对接口的假设

- 返回体默认带 `code`、`message`、`data`
- 列表接口默认返回 `data_total`
- 统计结果中的 `substat_dict` 使用编号字符串作为 key
- WebSocket 任意消息都可视为“需要刷新”

## 后续建议

- 将 URL 拼接从组件内抽离
- 用统一响应类型约束 `code/message/data`
- 为所有查询参数建立类型定义与默认值层
