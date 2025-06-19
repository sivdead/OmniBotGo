# Controller 层架构说明

## 整体结构

Controller 层采用分层设计，遵循整洁架构原则：

```
internal/controller/
├── http/                    # HTTP控制器
│   ├── middleware/          # 中间件
│   │   ├── logger.go       # 日志中间件
│   │   └── recovery.go     # 恢复中间件
│   ├── router.go           # 主路由配置
│   └── v1/                 # API v1版本
│       ├── bot.go          # 机器人管理
│       ├── channel.go      # 通道管理
│       ├── config.go       # 系统配置管理
│       ├── controller.go   # 核心控制器和消息处理
│       ├── error.go        # 错误处理
│       ├── health.go       # 健康检查
│       ├── log.go          # 日志管理
│       ├── monitor.go      # 监控统计
│       ├── platform.go     # 平台管理
│       ├── processor.go    # 处理器管理
│       ├── queue.go        # 队列管理
│       ├── response.go     # 响应处理
│       ├── router.go       # 路由定义
│       └── webhook.go      # Webhook处理
```

## 核心设计原则

### 1. 单一职责
每个控制器文件负责一个特定的业务领域：
- `bot.go` - 机器人的CRUD操作
- `channel.go` - 通道的CRUD操作和状态管理
- `message.go` - 消息处理和历史查询
- `webhook.go` - 平台消息回调处理

### 2. 统一响应格式
所有API使用统一的响应格式：
```go
type StandardResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Code    int         `json:"code,omitempty"`
}
```

### 3. 错误处理策略
- 参数验证错误：返回400状态码
- 资源不存在：返回404状态码
- 业务逻辑错误：返回500状态码
- 统一的错误日志记录

### 4. 依赖注入
控制器通过构造函数注入UseCase层依赖：
```go
type V1 struct {
    l logger.Interface
    v *validator.Validate
    messageUC MessageUseCase
    channelUC ChannelUseCase
    botUC     BotUseCase
}
```

## API 路由结构

### RESTful API 设计
```
/api/v1/
├── messages/               # 消息管理
│   ├── POST /send         # 发送消息
│   ├── GET /              # 获取消息历史
│   ├── GET /:id           # 获取消息详情
│   └── POST /:id/retry    # 重试失败消息
├── channels/              # 通道管理
│   ├── POST /             # 创建通道
│   ├── GET /              # 获取通道列表
│   ├── GET /:id           # 获取通道详情
│   ├── PUT /:id           # 更新通道
│   ├── DELETE /:id        # 删除通道
│   ├── PATCH /:id/status  # 更新通道状态
│   └── POST /:id/refresh-token # 刷新通道令牌
├── bots/                  # 机器人管理
├── logs/                  # 日志管理
├── configs/               # 系统配置
├── processors/            # 处理器管理
├── queues/                # 队列管理
├── platforms/             # 平台管理
└── monitor/               # 监控统计
```

### Webhook 路由
```
/webhook/:platform/:channel_id
├── GET    # 获取Webhook配置信息
└── POST   # 处理平台回调消息
```

### 健康检查路由
```
/healthz        # 简单健康检查
/health         # 详细健康检查
/ready          # 就绪状态检查
/live           # 存活状态检查
```

## 中间件功能

### 1. 日志中间件
- 记录所有HTTP请求的详细信息
- 包括IP、方法、URL、状态码、响应大小

### 2. 恢复中间件
- 捕获panic异常
- 返回统一的500错误响应
- 记录错误堆栈信息

### 3. 验证中间件
- 使用`validator`库进行参数验证
- 支持结构体标签验证
- 统一的验证错误响应格式

## 控制器特色功能

### 1. 消息控制器 (controller.go)
- 支持多种消息类型发送
- 消息历史查询with分页
- 时间范围过滤
- 消息状态和方向过滤
- 失败消息重试机制

### 2. Webhook控制器 (webhook.go)
- 支持多平台消息接收
- 签名验证机制
- URL验证处理
- 消息格式统一转换
- 平台适配器集成

### 3. 健康检查控制器 (health.go)
- 多层次健康检查
- 数据库连接状态
- 平台适配器状态
- 系统资源监控
- 启动时间统计

### 4. 监控控制器 (monitor.go)
- 系统概览信息
- 性能指标统计
- 时间序列数据
- 资源使用情况
- 详细健康状态

## 开发模式

### 1. 渐进式实现
当前采用渐进式开发模式：
- 先实现接口框架和路由
- 使用TODO注释标记未来实现点
- 返回临时响应保证API可用性
- 逐步替换为真实的UseCase调用

### 2. 测试友好设计
- 所有依赖通过接口注入
- 便于Mock和单元测试
- 明确的输入输出格式
- 统一的错误处理流程

### 3. 扩展性考虑
- 版本化API设计（v1目录）
- 可插拔的中间件架构
- 统一的响应格式
- 模块化的控制器结构

## 最佳实践

### 1. 错误处理
```go
if err := v.v.Struct(&req); err != nil {
    v.l.Error("请求验证失败", "error", err)
    return NewValidationErrorResponse(c, err)
}
```

### 2. 日志记录
```go
v.l.Info("消息发送成功", "message_id", message.MessageID, "channel_id", message.ChannelID)
```

### 3. 参数验证
```go
type SendMessageRequest struct {
    ChannelID   int64  `json:"channel_id" validate:"required"`
    MessageType string `json:"message_type" validate:"required"`
    Content     string `json:"content" validate:"required"`
}
```

### 4. 响应处理
```go
return NewSuccessResponse(c, http.StatusCreated, "创建成功", result)
```

这种架构设计确保了代码的可维护性、可测试性和可扩展性，同时提供了完整的API功能覆盖。 