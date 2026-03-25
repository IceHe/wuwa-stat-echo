# 前端数据库文档

## 结论

前端项目本身没有独立数据库，也没有本地持久化层。

它的数据来源全部是：

- 后端 HTTP API
- 后端 WebSocket 推送
- `src/stores/constants.ts` 中的静态常量

## 数据来源划分

### 1. 后端持久化数据

由后端 PostgreSQL 提供，前端只负责读写与展示：

- `wuwa_echo_log`
- `wuwa_tune_log`

前端通过接口访问这些数据，不直接访问数据库。

### 2. 前端静态领域数据

定义于 [src/stores/constants.ts](../src/stores/constants.ts)：

- 副词条类型定义
- 副词条档位定义
- 套装列表
- 套装颜色映射
- 共鸣者列表

这些数据相当于“前端内置字典”。

### 3. 前端运行时临时状态

保存在组件响应式对象中，例如：

- `echoLog`
- `template`
- `currentUser`
- `recentTuneStats`

这些状态刷新页面后会丢失，除非被同步到 URL 参数或重新从后端加载。

## 与数据库相关的前端注意点

- 前端的 `echoLog` 字段名基本直接映射后端 `EchoLog`
- 前端的统计页会直接消费后端聚合结果，而不是自行计算
- 页面中的“孔位”通常用 `0-4` 存储，但展示上对应第 `1-5` 孔
- `substat1..5` 是编码值，不是纯枚举编号

## 文档关联

- 后端数据库结构请见后端仓库中的 `docs/DATABASE.md`
- 前端字段说明请见 [DATA_FIELDS.md](./DATA_FIELDS.md)
