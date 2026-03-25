# 架构文档

## 项目概述

本项目是一个基于 FastAPI 的鸣潮（Wuthering Waves）游戏声骸调谐数据分析系统，用于记录、统计和分析玩家的声骸副词条调谐数据。

## 技术栈

- **Web 框架**: FastAPI
- **数据库**: PostgreSQL
- **ORM**: SQLModel
- **异步支持**: asyncio, httpx
- **WebSocket**: FastAPI WebSocket
- **服务器**: Uvicorn

## 项目结构

```
wuwa-fastapi/
├── api/                    # API 路由模块
│   ├── analysis.py        # 数据统计分析接口
│   ├── echo.py            # 声骸记录管理接口
│   ├── substat.py         # 副词条记录接口
│   ├── predict.py         # 预测相关接口
│   ├── db_data.py         # 数据库数据接口
│   ├── counts.py          # 统计计数接口
│   └── test.py            # 测试接口
├── db_init_data/          # 数据库初始化数据
├── main.py                # 应用入口
├── db.py                  # 数据库连接配置
├── model.py               # 数据模型定义
├── consts.py              # 常量定义（副词条、共鸣者模板等）
├── custom_types.py        # 自定义类型
├── response.py            # 响应格式定义
├── shared.py              # 共享状态
├── util.py                # 工具函数
├── ws.py                  # WebSocket 管理
└── db.sql                 # 数据库表结构
```

## 核心模块说明

### 1. 数据模型层 (model.py)

定义了两个核心数据表：

- **SubstatLog**: 副词条调谐记录表
  - 记录每次调谐的副词条类型、数值、位置等信息

- **EchoLog**: 声骸记录表
  - 记录完整声骸的所有副词条信息

### 2. 常量定义层 (consts.py)

- **SUBSTAT_DICT**: 副词条字典，包含 13 种副词条类型及其数值档位
- **RESONATOR_TEMPLATES**: 共鸣者模板，定义不同角色的评分权重
- **EXP**: 声骸升级经验表
- **调谐器回收率**: 资源回收相关常量

### 3. API 路由层

#### analysis.py

- `/tune_stats`: 副词条统计分析
- `/substat_distance_analysis`: 副词条间隔分析
- `/analyze_echo`: 声骸评分分析

#### echo.py

- `/echo_logs`: 声骸记录列表
- `/echo_log`: 创建/更新/查询声骸记录
- `/echo_log/{id}`: 获取单个声骸记录
- `/echo_log/find`: 查找相似声骸
- `/echo_logs/analysis`: 声骸数据分析

#### substat.py

- `/substat_logs`: 副词条记录列表
- `/tune_log`: 添加副词条记录
- `/tune_log/{id}/delete`: 删除副词条记录

#### counts.py

- `/counts/echo_dcrit`: 双暴声骸档位统计

### 4. 数据库层 (db.py)

使用 SQLModel 连接 PostgreSQL 数据库，提供 Session 依赖注入。

### 5. WebSocket 层 (ws.py)

提供实时数据推送功能，用于前端实时更新。

## 数据流

```
客户端请求
    ↓
FastAPI 路由
    ↓
业务逻辑处理
    ↓
SQLModel ORM
    ↓
PostgreSQL 数据库
    ↓
响应返回 / WebSocket 推送
```

## 核心算法

### 1. 副词条编码

使用位运算编码副词条信息：
- 低 13 位：副词条类型（0-12）
- 高位：数值档位（0-7）

### 2. 声骸评分算法

基于共鸣者模板计算声骸评分：
```
评分 = 主词条评分 + Σ(副词条数值 × 权重 / 最大分数 × 50)
```

### 3. 概率统计

- 副词条出现概率
- 特定档位概率
- 位置分布概率
- 双暴概率计算

## 部署架构

```
Uvicorn (ASGI Server)
    ↓
FastAPI Application
    ↓
PostgreSQL Database
```

启动命令：
```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8888
```

## 扩展性设计

1. **模块化路由**: 各功能模块独立，易于扩展
2. **依赖注入**: 使用 FastAPI 的依赖注入系统
3. **异步支持**: 全异步设计，支持高并发
4. **WebSocket**: 支持实时数据推送
5. **CORS 配置**: 支持跨域访问

## 性能优化

1. 使用 SQLModel 的连接池
2. 异步数据库操作
3. 位运算优化数据存储和计算
4. 缓存统计数据（shared.tune_stats）
