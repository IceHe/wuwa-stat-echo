# 鸣潮声骸调谐数据分析系统

一个基于 FastAPI 的鸣潮（Wuthering Waves）游戏声骸调谐数据记录与分析系统。

## 项目简介

本项目用于记录、统计和分析鸣潮游戏中声骸的副词条调谐数据，帮助玩家：

- 📊 记录每次调谐结果
- 📈 统计副词条真实概率
- 💎 评估声骸价值
- 🎯 分析资源消耗

## 主要功能

- **数据记录**: 记录声骸副词条调谐数据
- **概率统计**: 统计副词条出现概率和档位分布
- **声骸评分**: 根据角色需求评估声骸价值
- **资源分析**: 计算调谐器和经验消耗
- **实时推送**: WebSocket 实时数据更新

## 技术栈

- **Web 框架**: FastAPI
- **数据库**: PostgreSQL
- **ORM**: SQLModel
- **服务器**: Uvicorn
- **异步支持**: asyncio, httpx

## 快速开始

### 环境要求

- Python 3.13+
- PostgreSQL

### 安装依赖

```bash
pip install -r requirements.txt
```

### 配置数据库

先复制示例配置：

```bash
cp .env.example .env
```

再通过 `.env` 中的 `DATABASE_URL` 配置数据库连接；示例格式如下：

```python
postgresql://<user>:<password>@localhost:5432/<database>
```

`AUTH_SERVICE_URL` 与 `AUTH_SERVICE_TIMEOUT_SECONDS` 也可以一并写入 `.env`。

### 启动服务

```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8888
```

服务启动后访问：

- API 文档: http://localhost:8888/docs
- WebSocket: ws://localhost:8888/ws

### 启动前检查

- 应用启动时会在 lifespan 中执行 `init_tune_stats()` 预热统计缓存。
- 首次联调前建议先确认数据库中已有 `wuwa_tune_log` 与 `wuwa_echo_log` 两张表。

## 项目结构

```
wuwa-fastapi/
├── api/                    # API 路由模块
│   ├── analysis.py        # 数据统计分析
│   ├── echo.py            # 声骸记录管理
│   ├── substat.py         # 副词条记录
│   ├── counts.py          # 统计计数
│   └── ...
├── docs/                   # 文档目录
│   ├── ARCHITECTURE.md    # 架构文档
│   ├── API.md             # 接口文档
│   └── REQUIREMENTS.md    # 需求说明
├── main.py                # 应用入口
├── db.py                  # 数据库配置
├── model.py               # 数据模型
├── consts.py              # 常量定义
└── README.md              # 项目说明
```

## 文档

- [架构文档](docs/ARCHITECTURE.md) - 系统架构和技术设计
- [接口文档](docs/API.md) - API 接口详细说明
- [数据库文档](docs/DATABASE.md) - 表结构、索引与初始化说明
- [数据字段文档](docs/DATA_FIELDS.md) - 核心实体与统计返回字段
- [副词条文档](docs/ECHO_SUBSTATS.md) - 鸣潮声骸副词条与编码规则
- [需求说明](docs/REQUIREMENTS.md) - 功能需求和业务规则

## API 示例

### 获取副词条统计

```bash
curl http://localhost:8888/tune_stats
```

### 创建声骸记录

```bash
curl -X POST http://localhost:8888/echo_log \
  -H "Content-Type: application/json" \
  -d '{
    "substat1": 8193,
    "substat2": 8194,
    "s1_desc": "暴击 6.3%",
    "s2_desc": "暴击伤害 12.6%",
    "clazz": "沉日劫明",
    "user_id": 1
  }'
```

### 分析声骸评分

```bash
curl -X POST "http://localhost:8888/analyze_echo?resonator=暗主&cost=4C" \
  -H "Content-Type: application/json" \
  -d '{
    "substat1": 8193,
    "substat2": 8194,
    "substat_all": 3
  }'
```

## 支持的共鸣者

暗主、椿、珂莱塔、今汐、长离、坎特蕾拉、折枝、忌炎、相里要、洛可可、布兰特、菲比、赞妮、夏空、卡提希娅、露帕、弗洛洛、奥古斯塔、尤诺、嘉贝莉娜、陆赫斯、爱弥斯等 20+ 个角色。

## 开发计划

- [ ] 用户认证系统
- [ ] 数据可视化图表
- [ ] 批量导入导出
- [ ] 移动端适配
- [ ] 预测功能

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
