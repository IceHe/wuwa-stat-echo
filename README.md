# 鸣潮声骸调谐分析项目

这个仓库包含一个前后端分离的鸣潮声骸记录与分析工具：

- `backend/`：基于 FastAPI + SQLModel + PostgreSQL 的接口服务
- `frontend/`：基于 Vue 3 + Vite 的交互界面

项目主要用于记录声骸调谐结果、统计副词条分布、分析资源消耗，并按共鸣者模板给出声骸价值评估。

## 当前仓库结构

```text
.
├── backend/                # FastAPI 服务与数据库脚本
│   ├── api/                # 业务接口
│   ├── db_init_data/       # 初始化数据
│   ├── docs/               # 后端详细文档
│   ├── db.sql              # 表结构
│   └── main.py             # 服务入口
└── frontend/               # Vue 3 页面
    ├── src/components/     # 业务组件
    ├── src/views/          # 页面视图
    └── src/router/         # 路由配置
```

## 功能概览

- 声骸整条记录的新增、更新、删除、恢复与检索
- 单次副词条调谐日志记录
- 副词条概率、档位分布、位置分布统计
- 双暴与目标词条组合分析
- 基于共鸣者模板的声骸评分
- WebSocket 实时刷新部分统计结果

## 快速启动

### 1. 准备数据库

先在 `backend/` 目录下基于示例创建本地配置文件：

```bash
cp .env.example .env
```

然后在 `.env` 中填写 `DATABASE_URL`；示例格式为 `postgresql://<user>:<password>@localhost:5432/<database>`。

初始化表结构可参考 `backend/db.sql` 与 `backend/db_init_data/` 下的 SQL 文件。

### 2. 启动后端

在 `backend/` 目录下安装依赖并启动：

```bash
pip install -r requirements.txt
uvicorn main:app --reload --host 0.0.0.0 --port 8888
```

启动后默认接口地址：

- HTTP API：`http://127.0.0.1:8888`
- Swagger：`http://127.0.0.1:8888/docs`
- WebSocket：`ws://127.0.0.1:8888/ws`

### 3. 启动前端

在 `frontend/` 目录下：

```bash
npm install
npm run dev
```

默认开发地址：

- 前端页面：`http://127.0.0.1:3000`

## 鉴权接入

项目现在依赖 `~/wuwa/auth` 提供 token 鉴权。

- 前端新增 `/login` 页面，未登录时会先跳转登录。
- `echo` 后端会代理 `POST /auth/login` 与 `GET /auth/me` 到鉴权服务。
- 业务接口权限分级：
  - `view`：查询、统计、分析、WebSocket
  - `edit`：新增、修改、删除、恢复
  - `manage`：数据库修复类接口

后端默认通过以下环境变量访问鉴权服务：

- `AUTH_SERVICE_URL=http://127.0.0.1:8080`
- `AUTH_SERVICE_TIMEOUT_SECONDS=3`

这些变量也可以放在 `backend/.env` 中，由后端代码和 `systemd` 服务共同读取。

## 文档入口

- `backend/README.md`：后端说明与启动方式
- `backend/docs/ARCHITECTURE.md`：系统架构与模块划分
- `backend/docs/API.md`：主要 API 说明
- `backend/docs/REQUIREMENTS.md`：业务需求说明
- `backend/docs/ROADMAP_ADVANCED_STATS.md`：高级统计、决策支持与模拟器路线图
- `frontend/README.md`：前端页面结构与开发说明

## 适合后续补齐的文档

- 一份数据库初始化步骤文档
- 一份前后端联调与部署文档
