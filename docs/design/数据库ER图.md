# OmniBotGo 数据库 ER 图

## 📊 实体关系图

```mermaid
erDiagram
    %% 核心实体定义
    bots {
        bigint id PK
        varchar bot_name
        varchar bot_type
        text description
        varchar avatar_url
        json config
        tinyint status
        varchar created_by
        timestamp created_at
        timestamp updated_at
    }

    channels {
        bigint id PK
        bigint bot_id FK
        varchar platform_type
        varchar channel_name
        varchar webhook_path
        json config
        varchar access_token
        timestamp access_token_expires_at
        tinyint connection_status
        timestamp last_connected_at
        tinyint status
        timestamp created_at
        timestamp updated_at
    }

    messages {
        bigint id PK
        varchar message_id
        bigint channel_id FK
        varchar platform_message_id
        tinyint direction
        varchar message_type
        varchar content_type
        varchar sender_id
        varchar sender_name
        varchar sender_type
        varchar receiver_id
        varchar receiver_name
        varchar receiver_type
        text content
        json raw_content
        json unified_content
        varchar media_url
        varchar media_type
        bigint file_size
        tinyint message_status
        int retry_count
        text error_message
        bigint parent_message_id FK
        varchar conversation_id
        varchar backend_request_id
        timestamp platform_timestamp
        timestamp received_at
        timestamp processed_at
        timestamp sent_at
        timestamp created_at
        timestamp updated_at
    }

    message_processors {
        bigint id PK
        varchar processor_name
        varchar processor_type
        json config
        int priority
        tinyint status
        timestamp created_at
        timestamp updated_at
    }

    message_routing_rules {
        bigint id PK
        varchar rule_name
        text rule_description
        json bot_ids
        json platform_types
        json channel_ids
        json message_types
        json sender_patterns
        json content_patterns
        bigint processor_id FK
        tinyint route_type
        int priority
        tinyint is_fallback
        varchar condition_logic
        tinyint status
        timestamp created_at
        timestamp updated_at
    }

    system_configs {
        bigint id PK
        varchar config_key
        text config_value
        varchar config_type
        varchar config_group
        text description
        tinyint is_encrypted
        tinyint is_system
        timestamp created_at
        timestamp updated_at
    }

    message_queue {
        bigint id PK
        varchar queue_name
        varchar message_id
        json payload
        int priority
        int max_retries
        int retry_count
        tinyint status
        timestamp scheduled_at
        timestamp started_at
        timestamp completed_at
        text error_message
        timestamp created_at
        timestamp updated_at
    }

    connection_logs {
        bigint id PK
        bigint channel_id FK
        varchar event_type
        tinyint status
        varchar error_code
        text error_message
        json details
        timestamp created_at
    }

    api_call_logs {
        bigint id PK
        varchar request_id
        bigint channel_id FK
        bigint processor_id FK
        varchar method
        varchar url
        json request_headers
        text request_body
        int response_status
        json response_headers
        text response_body
        timestamp start_time
        timestamp end_time
        int duration_ms
        tinyint success
        text error_message
        timestamp created_at
    }

    %% 关系定义
    bots ||--o{ channels : "has"
    channels ||--o{ messages : "produces"
    channels ||--o{ connection_logs : "logs"
    channels ||--o{ api_call_logs : "calls"
    message_processors ||--o{ message_routing_rules : "routes_to"
    message_processors ||--o{ api_call_logs : "called_by"
    messages ||--o{ messages : "replies_to"
```

## 🔗 关系说明

### 核心关系
- **Bot → Channels**: 一个Bot可以配置多个平台通道（1:N）
- **Channels → Messages**: 一个通道产生多条消息（1:N）
- **Messages → Messages**: 消息可以回复其他消息（自关联）

### 路由关系
- **Message Processors → Routing Rules**: 一个消息处理器可以被多个路由规则使用（1:N）
- **Routing Rules**: 通过JSON字段匹配Bot、平台、通道等条件

### 日志关系
- **Channels → Connection Logs**: 记录通道连接状态变化（1:N）
- **Channels/Processors → API Call Logs**: 记录API调用详情（N:1）

## 📋 表类型分类

### 🤖 Bot管理层
- `bots` - Bot实例管理

### 🔌 连接层  
- `channels` - 平台通道配置

### 💬 消息层
- `messages` - 统一消息存储
- `message_queue` - 消息队列

### 🚀 处理层
- `message_processors` - 消息处理器配置
- `message_routing_rules` - 路由规则

### ⚙️ 系统层
- `system_configs` - 系统配置

### 📊 监控层
- `connection_logs` - 连接日志
- `api_call_logs` - API调用日志

## 🎯 架构特点

### 1. **层次清晰**
Bot实例 → 平台通道 → 消息流 → 消息处理器，每层职责明确

### 2. **Bot中心化**
所有资源围绕Bot实例组织，支持多Bot场景独立管理

### 3. **代码驱动**
平台适配器和处理器类型在代码中定义，数据库只存储实例配置

### 4. **智能路由**
基于多维度条件的消息路由，支持复杂业务场景

### 5. **监控完善**
连接状态、API调用、消息流转全程记录，支持问题诊断和性能优化

### 6. **扩展友好**
- 新增平台只需代码适配，无需数据库迁移
- JSON配置灵活适应不同平台需求
- 插件化的处理器架构 