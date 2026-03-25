# 数据库文档

## 概览

后端当前依赖 PostgreSQL，核心只使用两张业务表：

- `public.wuwa_echo_log`：完整声骸记录
- `public.wuwa_tune_log`：单次副词条调谐记录

初始化脚本位于 [db.sql](../db.sql)。

## 表设计

### 1. `wuwa_echo_log`

用途：存储一条完整声骸的最终副词条状态、套装、用户与时间信息。

主字段：

- `id`：自增主键
- `substat1` ~ `substat5`：5 个副词条编码值
- `substat_all`：副词条类型位图汇总，只关注低 13 位
- `s1_desc` ~ `s5_desc`：展示用文案，如 `暴击 8.1%`
- `clazz`：套装或分类名称
- `user_id`：用户 ID
- `deleted`：软删除标记
- `tuned_at`：调谐发生时间
- `created_at` / `updated_at`：记录创建与更新时间

查询特征：

- 列表接口按 `updated_at desc` 排序
- 相似声骸查询按 `substat1..5 + clazz + user_id + deleted` 组合过滤
- 分析接口会按 `user_id`、`id` 区间、`deleted` 等维度筛选

### 2. `wuwa_tune_log`

用途：存储每一次单孔调谐得到的副词条结果，作为统计样本来源。

主字段：

- `id`：自增主键
- `substat`：副词条类型编号，范围 `0-12`
- `value`：档位编号，范围 `0-7`
- `position`：孔位编号，范围 `0-4`
- `echo_id`：所属声骸 ID
- `user_id`：用户 ID
- `timestamp`：记录时间
- `deleted`：软删除标记

查询特征：

- 统计接口按 `id desc` 扫描最新样本
- 支持按 `user_id`、`after_id`、`before_id` 过滤
- 删除或恢复声骸时，会同步更新对应 `echo_id` 的调谐记录

## 约束

当前脚本中的校验约束：

- `wuwa_echo_log.deleted in (0, 1)`
- `wuwa_tune_log.substat between 0 and 12`
- `wuwa_tune_log.value between 0 and 7`
- `wuwa_tune_log.position between 0 and 4`
- `wuwa_tune_log.deleted in (0, 1)`

## 索引

### `wuwa_echo_log`

- `idx_wuwa_echo_log_deleted_updated_at`
  用于列表页与获取最新记录。
- `idx_wuwa_echo_log_user_id_deleted`
  用于按用户隔离查询。
- `idx_wuwa_echo_log_clazz_user_id_deleted`
  用于相似声骸和组合筛选。
- `idx_wuwa_echo_log_substat_all`
  用于目标词条位图过滤。
- `idx_wuwa_echo_log_substat_tuple`
  用于完整副词条组合匹配。

### `wuwa_tune_log`

- `idx_wuwa_tune_log_deleted_id`
  用于统计与日志列表的倒序扫描。
- `idx_wuwa_tune_log_user_id_deleted`
  用于用户维度统计。
- `idx_wuwa_tune_log_echo_id_position_deleted`
  用于按声骸与孔位删除/恢复。
- `idx_wuwa_tune_log_substat_deleted`
  用于副词条维度统计。

## 初始化与迁移建议

### 初始化

```sql
\i db.sql
```

### 当前实现特点

- 数据库脚本设计为可重复执行的 `CREATE TABLE IF NOT EXISTS`
- 当前项目没有正式迁移框架
- 结构变更主要通过手工维护 `db.sql`

### 后续建议

- 引入 Alembic 管理结构演进
- 为 `user_id + updated_at`、`deleted + tuned_at` 增加更贴近查询模式的索引评估
- 为历史大表增加归档或分区策略
