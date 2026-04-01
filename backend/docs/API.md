# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8888`
- **数据格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success message",
  "data": {}
}
```

### 错误响应
```json
{
  "code": -1,
  "message": "error message",
  "data": null
}
```

### 分页响应
```json
{
  "code": 0,
  "message": "page data",
  "data": [],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

## 1. 副词条统计接口

### 1.1 获取副词条统计数据

**接口**: `GET /tune_stats`

**描述**: 获取副词条的统计信息，包括出现次数、概率分布等

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| size | int | 否 | 限制查询数量，0 表示不限制 |
| user_id | int | 否 | 用户 ID，0 表示所有用户 |
| after_id | int | 否 | 查询 ID 大于此值的记录 |
| before_id | int | 否 | 查询 ID 小于此值的记录 |

**响应示例**:
```json
{
  "code": 0,
  "message": "tune stats",
  "data": {
    "data_total": 1000,
    "substat_dict": {
      "0": {
        "number": 0,
        "name": "Crit.Rate",
        "name_cn": "暴击",
        "total": 100,
        "percent": 10.0,
        "value_dict": {
          "0": {
            "value_number": 0,
            "value_desc": "6.3%",
            "total": 15,
            "percent": 15.0
          }
        }
      }
    },
    "substat_distance": [10, 5, 8, ...],
    "position_total": [200, 180, 150, 120, 100]
  }
}
```

### 1.2 副词条间隔分析

**接口**: `GET /substat_distance_analysis`

**描述**: 分析各副词条之间的出现间隔

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| size | int | 否 | 限制查询数量 |

---

## 2. 声骸记录接口

### 2.1 获取声骸记录列表

**接口**: `GET /echo_logs`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

**响应示例**:
```json
{
  "code": 0,
  "message": "echo logs",
  "data": [
    {
      "id": 1,
      "substat1": 8193,
      "substat2": 8194,
      "substat3": 0,
      "substat4": 0,
      "substat5": 0,
      "substat_all": 3,
      "s1_desc": "暴击 6.3%",
      "s2_desc": "暴击伤害 12.6%",
      "clazz": "沉日劫明",
      "user_id": 1,
      "deleted": 0,
      "tuned_at": "2025-03-20T10:30:00",
      "created_at": "2025-03-20T10:30:00",
      "updated_at": "2025-03-20T10:30:00"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

### 2.2 创建声骸记录

**接口**: `POST /echo_log`

**请求体**:
```json
{
  "substat1": 8193,
  "substat2": 8194,
  "substat3": 0,
  "substat4": 0,
  "substat5": 0,
  "s1_desc": "暴击 6.3%",
  "s2_desc": "暴击伤害 12.6%",
  "clazz": "沉日劫明",
  "user_id": 1,
  "tuned_at": "2025-03-20T10:30:00"
}
```

### 2.3 更新声骸记录

**接口**: `PATCH /echo_log`

**请求体**: 同创建接口，需包含 `id` 字段

### 2.4 删除声骸记录

**接口**: `DELETE /echo_log/{id}`

**说明**: 软删除，将 `deleted` 字段设为 1

### 2.5 恢复声骸记录

**接口**: `POST /echo_log/{id}/recover`

### 2.6 获取单个声骸记录

**接口**: `GET /echo_log/{id}`

**说明**:
- id > 0: 获取指定 ID 的记录
- id = 0 或 -1: 获取最新的未删除记录

**响应示例**:
```json
{
  "code": 0,
  "message": "echo log",
  "data": {
    "id": 1,
    "substat1": 8193,
    "pos_total": 150
  }
}
```

### 2.7 查找相似声骸

**接口**: `POST /echo_log/find`

**描述**: 根据玩家 ID、声骸 ID、套装、关键词和副词条组合搜索声骸记录，可搜索其他用户录入的数据

**请求体**:
```json
{
  "id": 12345,
  "keyword": "暴击",
  "substat1": 8193,
  "substat2": 8194,
  "user_id": 1,
  "clazz": "沉日劫明"
}
```

**说明**:
- 至少提供一个搜索条件，否则返回空列表
- `keyword` 支持匹配玩家 ID、声骸 ID、套装名和 `s1_desc` 到 `s5_desc` 的副词条描述

### 2.8 声骸数据分析

**接口**: `GET /echo_logs/analysis`

**描述**: 分析声骸数据，计算目标副词条的出现频率、平均间隔等

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | int | 否 | 用户 ID |
| size | int | 否 | 限制查询数量 |
| target_bits | int | 否 | 目标副词条位掩码，默认 0b11（双暴） |
| after_echo_id | int | 否 | 起始声骸 ID |
| before_echo_id | int | 否 | 结束声骸 ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "echo logs analysis",
  "data": {
    "target_echo_distance": 10,
    "target_substat_distance": 35,
    "target": 5,
    "target_avg_echo": 20.0,
    "target_avg_substat": 70.0,
    "tuner_consumed": 500,
    "tuner_consumed_avg": 100,
    "exp_consumed": 50,
    "exp_consumed_avg": 10
  }
}
```

---

## 3. 副词条记录接口

### 3.1 获取副词条记录列表

**接口**: `GET /substat_logs`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

### 3.2 添加副词条记录

**接口**: `POST /tune_log`

**请求体**:
```json
{
  "user_id": 1,
  "echo_id": 1,
  "position": 1,
  "substat": 0,
  "value": 3
}
```

**字段说明**:
- `position`: 副词条位置（1-5）
- `substat`: 副词条类型（0-12）
- `value`: 数值档位（0-7）

### 3.3 删除副词条记录

**接口**: `POST /tune_log/{id}/delete`

### 3.5 删除声骸指定位置的副词条

**接口**: `DELETE /echo_log/{echoId}/substat_pos/{pos}`

---

## 4. 声骸评分接口

### 4.1 分析声骸评分

**接口**: `POST /analyze_echo`

**描述**: 根据共鸣者模板计算声骸评分

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| resonator | string | 否 | 共鸣者名称（如"暗主"、"椿"等） |
| cost | string | 否 | Cost 类型（如"4C"、"3C属伤"、"1C"），默认"1C" |

**请求体**:
```json
{
  "substat1": 8193,
  "substat2": 8194,
  "substat3": 0,
  "substat4": 0,
  "substat5": 0,
  "substat_all": 3
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "echo log",
  "data": {
    "score": {
      "resonator": "暗主",
      "substat1": 8.5,
      "substat2": 7.2,
      "substat_all": 25.8
    },
    "two_crit_percent": 33.33,
    "substat_dict": {...}
  }
}
```

---

## 5. 统计计数接口

### 5.1 双暴声骸档位统计

**接口**: `GET /counts/echo_dcrit`

**描述**: 统计所有双暴声骸（同时拥有暴击和暴击伤害）的档位分布情况

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| size | int | 否 | 限制查询数量，0 表示不限制 |
| after_id | int | 否 | 查询 ID 大于此值的记录 |
| before_id | int | 否 | 查询 ID 小于此值的记录 |

**响应示例**:
```json
{
  "code": 0,
  "message": "test",
  "data": {
    "echo_count": 1000,
    "dcrit_total": 150,
    "counts": {
      "0": {
        "0": 5,
        "1": 8,
        "2": 10,
        "3": 12,
        "4": 8,
        "5": 6,
        "6": 4,
        "7": 2
      },
      "1": {
        "0": 6,
        "1": 9,
        "2": 11,
        "3": 10,
        "4": 7,
        "5": 5,
        "6": 3,
        "7": 1
      }
    }
  }
}
```

**字段说明**:
- `echo_count`: 总声骸数量
- `dcrit_total`: 双暴声骸总数
- `counts`: 档位分布统计
  - 第一层 key: 暴击档位（0-7）
  - 第二层 key: 暴击伤害档位（0-7）
  - value: 该档位组合的数量

**档位对应数值**:

暴击档位：
- 0: 6.3%
- 1: 6.9%
- 2: 7.5%
- 3: 8.1%
- 4: 8.7%
- 5: 9.3%
- 6: 9.9%
- 7: 10.5%

暴击伤害档位：
- 0: 12.6%
- 1: 13.8%
- 2: 15.0%
- 3: 16.2%
- 4: 17.4%
- 5: 18.6%
- 6: 19.8%
- 7: 21.0%

**使用场景**:
- 分析双暴声骸的档位分布
- 评估获得高档位双暴的概率
- 统计不同档位组合的出现频率

---

## 6. WebSocket 接口

### 6.1 实时数据推送

**接口**: `WS /ws`

**描述**: 建立 WebSocket 连接，接收实时数据更新

**消息格式**:
```json
{
  "type": "update_echo_log",
  "data": {}
}
```

---

## 副词条编码说明

副词条使用位运算编码：

### 副词条类型（低 13 位）
- 0: 暴击
- 1: 暴击伤害
- 2: 攻击
- 3: 防御
- 4: 生命
- 5: 攻击固定值
- 6: 防御固定值
- 7: 生命固定值
- 8: 共鸣效率
- 9: 普攻
- 10: 重击
- 11: 共鸣技能
- 12: 共鸣解放

### 数值档位（高位）
每种副词条有 4-8 个档位，档位编号从 0 开始

### 编码示例
```
暴击 6.3% = (1 << 0) | (1 << (13 + 0)) = 8193
暴击伤害 12.6% = (1 << 1) | (1 << (13 + 0)) = 8194
```

---

## 共鸣者模板

支持的共鸣者：
- 通用
- 暗主
- 椿
- 珂莱塔
- 今汐
- 长离
- 坎特蕾拉
- 折枝
- 忌炎
- 相里要
- 洛可可
- 布兰特
- 菲比
- 赞妮
- 夏空
- 卡提希娅
- 露帕
- 弗洛洛
- 奥古斯塔
- 尤诺
- 嘉贝莉娜
- 陆赫斯
- 爱弥斯

每个共鸣者有不同的副词条权重和评分标准。
