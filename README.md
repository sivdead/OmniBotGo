# **OmniBotGo - 统一消息适配器网关**

基于Go语言开发的高性能、可扩展的统一消息适配器网关，能够**单一实例同时**连接和管理多个即时通讯平台（如企业微信、钉钉、公众号等），并将消息中继到后端业务逻辑或AI服务，实现消息的双向传递。

[![License](https://img.shields.io/badge/License-MIT-success)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/sivdead/OmniBotGo)](https://goreportcard.com/report/github.com/sivdead/OmniBotGo)

[![Web Framework](https://img.shields.io/badge/Fiber-Web%20Framework-blue)](https://github.com/gofiber/fiber)
[![Dependency Injection](https://img.shields.io/badge/Wire-Compile--time%20DI-blue)](https://github.com/google/wire)
[![API Documentation](https://img.shields.io/badge/Swagger-API%20Documentation-blue)](https://github.com/swaggo/swag)
[![ORM](https://img.shields.io/badge/GORM-Database%20ORM-blue)](https://gorm.io/)
[![Database](https://img.shields.io/badge/MySQL-Database-blue)](https://www.mysql.com/)
[![Validation](https://img.shields.io/badge/Validator-Data%20Integrity-blue)](https://github.com/go-playground/validator)
[![JSON Handling](https://img.shields.io/badge/Go--JSON-Fast%20Serialization-blue)](https://github.com/goccy/go-json)
[![Database Migrations](https://img.shields.io/badge/Migrations-Seamless%20Schema%20Updates-blue)](https://github.com/golang-migrate/migrate)
[![Logging](https://img.shields.io/badge/ZeroLog-Structured%20Logging-blue)](https://github.com/rs/zerolog)
[![Metrics](https://img.shields.io/badge/Prometheus-Metrics%20Integration-blue)](https://github.com/ansrivas/fiberprometheus)
[![Testing](https://img.shields.io/badge/Testify-Testing%20Framework-blue)](https://github.com/stretchr/testify)
[![Mocking](https://img.shields.io/badge/Mock-Mocking%20Library-blue)](https://go.uber.org/mock)

## 项目目标

构建一个统一的消息适配器网关，实现：
- **多平台并发管理**：在单个实例中同时连接多个IM平台
- **消息双向中继**：将平台消息路由到后端服务，并将响应消息返回到目标平台
- **高性能与稳定性**：支持大量并发消息处理和7x24小时稳定运行
- **可扩展性**：支持轻松添加新的平台适配器

## 核心功能

### 🚀 核心引擎 (P0)
- **多平台并发管理**：单实例同时管理多个IM平台连接
- **统一消息模型**：平台无关的标准消息格式
- **智能消息路由**：基于来源和目标的精确消息分发
- **后端服务对接**：HTTP/S Webhook + API方式的双向通信
- **优雅启停与重连**：平滑运行和自动恢复机制
- **编译时依赖注入**：基于Google Wire的类型安全依赖管理

### 📱 支持的平台
- **企业微信 (WeCom)**：应用消息收发、多种消息类型、事件处理
- **钉钉 (DingTalk)**：机器人消息、企业内部应用支持
- **微信公众号**：订阅号和服务号消息处理
- **可扩展架构**：预留接口支持飞书、Telegram、Slack等

### 💬 消息类型支持
- **基础消息**：文本、图片、语音、视频
- **富媒体消息**：Markdown、链接、文件、卡片
- **事件消息**：关注/取消、进群/退群、菜单点击等

## 快速开始

### 本地开发

```sh
# 启动 MySQL 和 RabbitMQ
just compose-up

# 运行应用（包含依赖注入代码生成和数据库迁移）
just run
```

### 集成测试

```sh
# 启动完整测试环境
just compose-up-integration-test
```

### 完整 Docker 部署

```sh
# 启动完整服务栈（含反向代理）
just compose-up-all 
```

### 服务检查

启动后可访问以下服务：

- **REST API**:
  - http://app.lvh.me/healthz | http://127.0.0.1:8080/healthz
  - http://app.lvh.me/metrics | http://127.0.0.1:8080/metrics  
  - http://app.lvh.me/swagger | http://127.0.0.1:8080/swagger

- **gRPC 服务**:
  - URL: `tcp://grpc.lvh.me:8081` | `tcp://127.0.0.1:8081`

- **MySQL 数据库**:
  - `mysql://user:myAwEsOm3pa55@w0rd@127.0.0.1:3306/omnibotgo`

- **RabbitMQ**:
  - http://rabbitmq.lvh.me | http://127.0.0.1:15672
  - 用户名/密码: `guest` / `guest`

## 项目架构

本项目基于整洁架构原则，采用依赖倒置设计，使用Google Wire进行编译时依赖注入：

### 依赖注入架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Config        │    │   Logger        │    │   Database      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┼───────────┐
                                 │                      │           │
                          ┌──────▼──────┐    ┌─────────▼──────┐    │
                          │TranslationWA│    │TranslationRepo │    │
                          │     API     │    │               │    │
                          └──────┬──────┘    └─────────┬──────┘    │
                                 │                     │           │
                                 └─────────┬───────────┘           │
                                           │                       │
                                  ┌────────▼────────┐              │
                                  │TranslationUseCase│              │
                                  └────────┬────────┘              │
                                           │                       │
                      ┌────────────────────┼─────────────────────┬─┘
                      │                    │                     │
              ┌───────▼────┐    ┌─────────▼────┐    ┌──────────▼────┐
              │HTTPServer  │    │GRPCServer    │    │RMQServer      │
              └───────┬────┘    └─────────┬────┘    └──────────┬────┘
                      │                   │                    │
                      └───────────────────┼────────────────────┘
                                          │
                                  ┌───────▼────────┐
                                  │       App      │
                                  └────────────────┘
```

**Wire的优势**：
- **编译时安全**：依赖错误在编译期发现
- **无运行时开销**：生成普通Go代码，无反射
- **类型安全**：基于Go类型系统的依赖管理
- **易于测试**：支持Mock依赖轻松注入

### 目录结构

```
OmniBotGo/
├── cmd/app/                 # 应用程序入口
├── config/                  # 配置文件目录（YAML格式）
├── internal/
│   ├── app/                # 应用启动和Wire依赖注入
│   ├── config/             # 配置管理逻辑（Viper）
│   ├── controller/         # 控制器层（HTTP/gRPC/AMQP）
│   ├── entity/             # 业务实体和消息模型
│   ├── providers/          # Wire Provider 函数定义
│   ├── usecase/            # 业务逻辑层
│   └── repo/               # 数据访问层抽象
├── pkg/                    # 公共工具包
│   ├── mysql/              # MySQL 连接包
│   ├── httpserver/         # HTTP 服务器包
│   ├── grpcserver/         # gRPC 服务器包
│   └── rabbitmq/           # RabbitMQ RPC 包
├── migrations/             # 数据库迁移文件
├── docs/                   # 项目文档和 API 文档
├── README_CONFIG.md        # 配置系统详细说明
└── integration-test/       # 集成测试
```

### 核心组件

#### `cmd/app/main.go`
应用程序启动入口，负责配置初始化和调用Wire依赖注入。

#### `internal/app/`
应用程序生命周期管理：
- **wire.go**: Wire依赖注入配置和Injector函数
- **wire_gen.go**: Wire自动生成的依赖注入代码
- **app.go**: 简化的应用启动逻辑

#### `internal/providers/`
Wire Provider函数定义，采用模块化设计：
- **infrastructure.go**: 基础设施Provider（Logger、Database）
- **repository.go**: 数据访问层Provider
- **usecase.go**: 业务逻辑层Provider
- **servers.go**: 服务器层Provider（HTTP、gRPC、RMQ）
- **app.go**: 应用程序主结构体Provider

#### `internal/config/` & `config/`
基于Viper的配置管理系统：
- **`internal/config/`**: 配置管理逻辑、类型定义、验证规则
- **`config/`**: YAML配置文件存储目录
- **多源配置**: 支持环境变量覆盖配置文件
- **类型安全**: 强类型配置结构体，编译时检查
- **配置验证**: 启动时验证必需配置项

详细说明请参考 [README_CONFIG.md](README_CONFIG.md)

#### `internal/controller/`
多协议服务器支持：
- **HTTP REST API** (基于 Fiber 框架)
- **gRPC 服务** (基于 protobuf)  
- **AMQP RPC** (基于 RabbitMQ)

#### `internal/entity/`
统一消息模型和业务实体定义，提供平台无关的消息格式。

#### `internal/usecase/`
核心业务逻辑层：
- 消息路由和转换
- 平台适配器管理
- 后端服务集成
- 事件处理逻辑

#### `internal/repo/`
数据访问层，支持：
- **MySQL 数据库**（GORM + Squirrel 混合架构）
- **外部 WebAPI** 调用
- **缓存和队列** 服务

**数据库访问架构说明**：
- **GORM为主**：处理模型定义、CRUD操作、关系映射、事务管理
- **Squirrel为辅**：处理复杂查询、动态SQL构建、统计分析、性能优化场景
- **共享连接池**：两者使用同一个MySQL连接池，避免资源浪费
- **类型安全**：GORM提供编译时类型检查，Squirrel提供运行时灵活性

#### `pkg/mysql/`
MySQL 数据库连接管理：
- 连接池管理
- 自动重连机制
- GORM ORM 集成
- Squirrel 查询构建器集成
- 事务支持

## 技术栈

### 核心技术
- **语言**: Go 1.24+
- **Web框架**: Fiber v2（高性能 HTTP 框架）
- **依赖注入**: Google Wire（编译时依赖注入）
- **数据库**: MySQL 8.0+
- **ORM**: GORM（Go语言最受欢迎的ORM）
- **查询构建器**: Squirrel（用于复杂SQL查询）
- **消息队列**: RabbitMQ
- **协议**: HTTP/HTTPS, gRPC, AMQP

### 基础设施
- **配置管理**: 环境变量 + YAML/TOML
- **日志**: Zerolog（结构化日志）
- **监控**: Prometheus + Grafana
- **文档**: Swagger/OpenAPI 3.0
- **容器**: Docker + Docker Compose
- **CI/CD**: GitHub Actions

### 开发工具
- **代码质量**: golangci-lint
- **依赖注入代码生成**: Wire
- **测试**: Testify + 集成测试
- **Mock**: go-mock
- **迁移**: golang-migrate

## 配置说明

### 环境变量配置

主要环境变量：

```bash
# 应用基础配置
APP_NAME=omnibotgo
APP_VERSION=1.0.0

# HTTP 服务配置
HTTP_PORT=8080
HTTP_USE_PREFORK_MODE=false

# MySQL 数据库配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=omnibotgo
MYSQL_USERNAME=root
MYSQL_PASSWORD=password
MYSQL_MAX_CONNECTIONS=100

# gRPC 服务配置
GRPC_PORT=8081

# RabbitMQ 配置
RMQ_URL=amqp://guest:guest@localhost:5672/
RMQ_RPC_SERVER=rpc_server
RMQ_RPC_CLIENT=rpc_client

# 平台配置示例
# 企业微信
WECOM_CORP_ID=your_corp_id
WECOM_APP_SECRET=your_app_secret

# 钉钉
DINGTALK_APP_KEY=your_app_key
DINGTALK_APP_SECRET=your_app_secret

# 公众号
WECHAT_APPID=your_appid
WECHAT_APP_SECRET=your_app_secret

# 监控和调试
LOG_LEVEL=info
METRICS_ENABLED=true
SWAGGER_ENABLED=true
```

### 平台适配器配置

每个平台都有独立的配置节，支持多账号配置：

```yaml
platforms:
  wecom:
    - name: "default"
      corp_id: "xxxx"
      app_secret: "xxxx"
      token: "xxxx"
      encoding_aes_key: "xxxx"
      
  dingtalk:
    - name: "default"
      app_key: "xxxx"
      app_secret: "xxxx"
      
  wechat_official:
    - name: "service_account"
      appid: "xxxx"
      app_secret: "xxxx"
      token: "xxxx"
      encoding_aes_key: "xxxx"
```

## API 文档

### REST API

启动服务后访问 `/swagger` 端点查看完整的API文档。

主要端点：
- `GET /healthz` - 健康检查
- `GET /metrics` - Prometheus 指标
- `POST /webhook/{platform}` - 平台消息接收
- `POST /api/v1/message/send` - 消息发送
- `GET /api/v1/message/history` - 消息历史

### gRPC API

protobuf 定义文件位于 `docs/proto/` 目录。

### 消息格式

#### 统一消息格式
```json
{
  "id": "msg_123456",
  "platform": "wecom",
  "platform_msg_id": "platform_specific_id",
  "from": {
    "platform_user_id": "user123",
    "username": "张三",
    "user_type": "user"
  },
  "to": {
    "platform_group_id": "group456",
    "group_name": "开发群",
    "target_type": "group"
  },
  "message_type": "text",
  "content": {
    "text": "Hello World"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "context": {
    "conversation_id": "conv_789",
    "reply_to_msg_id": "msg_456"
  }
}
```

## 开发指南

### 开发环境设置

1. **安装依赖**
```bash
go mod download
just bin-deps  # 安装开发工具
```

2. **启动开发环境**
```bash
just compose-up  # 启动 MySQL 和 RabbitMQ
just run        # 运行应用
```

3. **代码质量检查**
```bash
just linter-golangci  # 代码检查
just test            # 运行测试
just format          # 代码格式化
just wire           # 生成依赖注入代码
```

### 添加新的平台适配器

1. **定义平台接口**
在 `internal/usecase/platform/` 下创建平台接口定义

2. **实现适配器**
在 `internal/adapter/` 下实现具体的平台适配器

3. **创建Wire Provider**
在 `internal/providers/` 下添加新平台的Provider函数：
```go
func NewWeChatAdapter(cfg *config.Config) (*wechat.Adapter, error) {
    return wechat.New(cfg.WeChat)
}
```

4. **更新ProviderSet**
将新的Provider添加到相应的ProviderSet中

5. **重新生成Wire代码**
```bash
just wire  # 或直接运行 wire ./internal/app
```

6. **添加消息映射**
在 `internal/entity/message.go` 中添加平台特定的消息类型映射

### 数据库迁移

```bash
# 创建新的迁移文件
just migrate-create migration_name

# 执行迁移
just migrate-up

# 回滚迁移  
migrate -path migrations -database 'mysql://user:pass@localhost:3306/omnibotgo' down 1
```

### 测试

```bash
# 单元测试
just test

# 集成测试
just integration-test

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 数据库使用示例

项目采用 **GORM + Squirrel** 混合架构，各司其职：

#### GORM 使用场景
```go
// 简单 CRUD 操作
func (r *TranslationRepo) Store(ctx context.Context, t entity.Translation) error {
    return r.DB.WithContext(ctx).Create(&t).Error
}

// 批量操作
func (r *TranslationRepo) BatchInsert(ctx context.Context, items []entity.Translation) error {
    return r.DB.WithContext(ctx).CreateInBatches(items, 100).Error
}

// 关联查询
func (r *TranslationRepo) GetWithUser(ctx context.Context, id uint) (*entity.Translation, error) {
    var translation entity.Translation
    err := r.DB.WithContext(ctx).Preload("User").First(&translation, id).Error
    return &translation, err
}
```

#### Squirrel 使用场景
```go
// 复杂动态查询
func (r *TranslationRepo) SearchWithConditions(ctx context.Context, conditions map[string]interface{}) ([]entity.Translation, error) {
    query := r.Builder.Select("*").From("translations")
    
    // 动态添加条件
    if source, ok := conditions["source"]; ok {
        query = query.Where("source = ?", source)
    }
    if keyword, ok := conditions["keyword"]; ok {
        query = query.Where("original LIKE ?", "%"+keyword.(string)+"%")
    }
    
    sql, args, _ := query.ToSql()
    rows, err := r.SqlDB.QueryContext(ctx, sql, args...)
    // ... 处理结果
}

// 统计分析查询
func (r *TranslationRepo) GetStatistics(ctx context.Context) ([]map[string]interface{}, error) {
    sql, args, _ := r.Builder.
        Select("source", "destination", "COUNT(*) as count", "AVG(LENGTH(original)) as avg_length").
        From("translations").
        GroupBy("source", "destination").
        OrderBy("count DESC").
        ToSql()
        
    rows, err := r.SqlDB.QueryContext(ctx, sql, args...)
    // ... 处理结果
}
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t omnibotgo:latest .

# 使用 Docker Compose 部署
docker-compose up -d
```

### Kubernetes 部署

部署配置文件位于 `deployments/k8s/` 目录：

```bash
kubectl apply -f deployments/k8s/
```

### 监控和日志

- **指标监控**: Prometheus + Grafana
- **日志聚合**: ELK Stack 或 Loki
- **分布式追踪**: Jaeger 或 Zipkin
- **健康检查**: `/healthz` 端点

## 贡献指南

1. **Fork 本仓库**
2. **创建功能分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'Add some AmazingFeature'`)
4. **推送到分支** (`git push origin feature/AmazingFeature`)
5. **打开 Pull Request**

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `golangci-lint` 进行代码检查
- 编写单元测试，保持测试覆盖率 > 80%
- 提交前运行 `just pre-commit`

## 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 联系方式

- **项目主页**: github.com/sivdead/OmniBotGo
- **问题反馈**: github.com/sivdead/OmniBotGo/issues
- **讨论社区**: github.com/sivdead/OmniBotGo/discussions

## 致谢

- 基于 [go-clean-template](https://github.com/sivdead/OmniBotGo) 构建
- 感谢所有贡献者的支持

---

⭐ 如果这个项目对你有帮助，请给它一个星标！
