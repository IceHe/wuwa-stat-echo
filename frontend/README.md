# 前端说明

前端是一个基于 Vue 3 + TypeScript + Vite 的鸣潮声骸分析页面，用于操作后端接口并展示调谐统计、声骸详情、目标词条分析等结果。

## 技术栈

- Vue 3
- TypeScript
- Vue Router
- Axios
- Vite

## 目录结构

```text
frontend/
├── docs/                   # 前端文档
├── src/components/         # 业务组件
├── src/views/              # 页面视图
├── src/router/             # 路由
├── src/stores/constants.ts # 常量与后端地址配置
└── vite.config.ts          # Vite 开发配置
```

## 页面与功能

当前路由主要包括：

- `/echo`：声骸记录与编辑主页面
- `/substat`：副词条记录与统计页面
- `/analysis`：分析相关页面
- `/echo_board`：目标词条与局部分析页面
- `/echo_dcrit_count`：双暴统计页面

默认 `/` 会直接进入 `EchoView`。

## 本地开发

安装依赖：

```sh
npm install
```

启动开发服务器：

```sh
npm run dev
```

默认监听：

- `0.0.0.0:3000`

构建生产包：

```sh
npm run build
```

代码检查：

```sh
npm run lint
```

## 后端联调

前端当前根据浏览器访问的主机名自动拼接后端地址，逻辑在 `src/stores/constants.ts`：

```ts
const API_PORT = '8888'
const API_HOST = typeof window === 'undefined' ? '127.0.0.1' : window.location.hostname
export const API_SERV = `${API_HOST}:${API_PORT}`
```

页面中同时会访问：

- HTTP 接口：`http://${API_SERV}/...`
- WebSocket：`ws://${API_SERV}/ws`

因此前后端地址必须保持一致。

## 文档

- [架构文档](docs/ARCHITECTURE.md)
- [接口文档](docs/API.md)
- [数据库文档](docs/DATABASE.md)
- [数据字段文档](docs/DATA_FIELDS.md)
- [副词条文档](docs/ECHO_SUBSTATS.md)

## 现状说明

- 当前项目更偏向内部使用工具，配置方式比较直接。
- 如果后续继续维护，优先建议把 API 地址迁移到 Vite 环境变量，例如 `VITE_API_BASE_URL`。
