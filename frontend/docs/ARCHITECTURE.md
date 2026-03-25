# 前端架构文档

## 概览

前端是一个基于 Vue 3 + Vite 的单页应用，目标是把后端的声骸记录、调谐日志、统计分析与目标词条分析能力组织成可操作页面。

技术栈：

- Vue 3
- Vue Router
- Axios
- 原生 WebSocket
- Vite

## 页面结构

路由定义位于 [src/router/index.ts](../src/router/index.ts)。

| 路由 | 视图 | 作用 |
|---|---|---|
| `/` | `EchoView` | 默认首页，完整声骸记录与日志联动 |
| `/home` | `HomeView` | 早期副词条记录与统计页 |
| `/substat` | `SubstatView` | 单孔调谐记录与统计 |
| `/echo` | `EchoView` | 声骸编辑主页面 |
| `/analysis` | `AnalysisView` | 分析类页面 |
| `/echo_board` | `EchoBoardView` | 目标词条与单条声骸分析 |
| `/echo_dcrit_count` | `EchoDcritCountView` | 双暴统计页 |

## 组件分层

### 1. 页面视图层

位于 `src/views/`，负责把多个组件拼成完整工作台。

示例：

- `EchoView.vue`
  组合 `Echo`、`EchoLogs`、`FindEcho`、`SubstatLogs`
- `SubstatView.vue`
  组合单孔记录面板和统计面板

### 2. 业务组件层

位于 `src/components/`，负责具体交互逻辑：

- `Echo.vue`：编辑当前声骸、记录副词条、分析目标词条
- `Substat.vue`：录入单条调谐记录
- `EchoLogs.vue` / `SubstatLogs.vue`：历史列表
- `SubstatAnalysis.vue`：副词条统计
- `FindEcho.vue`：按词条组合查找声骸
- `EchoBoard.vue`：围绕单条声骸进行分析
- `EchoDcritCount.vue`：双暴档位统计

### 3. 常量与共享配置层

位于 `src/stores/constants.ts`，负责维护：

- 后端地址
- 套装常量
- 共鸣者列表
- 副词条与档位常量
- 颜色映射

## 数据流

```text
用户操作
  ↓
Vue 组件状态变更
  ↓
Axios / WebSocket
  ↓
FastAPI 后端
  ↓
PostgreSQL 统计结果
  ↓
组件重新渲染
```

## 与后端的耦合点

前端与后端耦合度较高，主要体现在：

- 直接依赖后端字段命名，如 `substat1`、`substat_all`、`clazz`
- 直接拼接 HTTP 路径，没有封装独立 API SDK
- 依赖后端统计返回结构，如 `substat_dict`、`position_total`
- 使用 WebSocket `/ws` 接收刷新广播

## 运行时配置

后端主机名通过当前浏览器访问主机自动推导：

```ts
const API_HOST = window.location.hostname
const API_SERV = `${API_HOST}:8888`
```

这意味着：

- 访问 `http://172.31.0.2:3000` 时，前端会请求 `http://172.31.0.2:8888`
- 前后端更适合部署在同一台主机或同一入口域名下

## 当前架构特点

优点：

- 页面直连接口，开发成本低
- 常量集中，副词条显示规则统一
- 统计与日志列表联动简单直接

限制：

- 组件偏大，业务逻辑和展示耦合
- 类型定义不足，部分数据对象仍依赖隐式结构
- API 访问未抽象，后续维护成本较高
- 缺少环境变量与多环境配置

## 后续建议

- 把接口访问抽到 `src/api/`
- 为 `EchoLog`、`SubstatLog`、`TuneStats` 建正式 TypeScript 类型
- 将副词条常量拆分为独立领域模块
- 为页面添加统一错误提示和加载状态
