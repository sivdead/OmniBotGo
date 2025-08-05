# 项目架构指南

## 整洁架构设计

本项目基于整洁架构原则，采用依赖倒置设计：

### 核心原则
- 依赖方向：Controller → UseCase → Entity ← Repository
- 业务逻辑独立于框架、数据库、外部服务
- **演进式设计 (Evolutionary Design)**: 优先实现当前需求，保持设计的灵活性，避免为未来不确定的需求进行过度工程。
- **实用主义与简洁性 (Pragmatism & Simplicity)**: 选择最简单、直接的方案解决问题，避免不必要的复杂性。遵循YAGNI (You Ain't Gonna Need It)原则。
- **避免过度设计 (Avoid Over-engineering)**: 警惕设计模式的滥用和不成熟的抽象。优先编写清晰、可读的代码，在需要时进行重构。
- 使用接口进行抽象，便于测试和扩展

### 目录结构
```
cmd/app/                    # 应用程序入口点
internal/
├── app/                    # 应用启动和依赖注入
├── controller/             # 控制器层 (HTTP/gRPC/AMQP)
├── usecase/                # 业务逻辑层
├── entity/                 # 业务实体和领域模型
├── repo/                   # 数据访问层接口
└── providers/              # 依赖注入Provider函数
pkg/                        # 公共基础设施包
```

### 依赖注入
项目使用 Google Wire 进行编译时依赖注入：
- Wire配置在 `internal/app/wire.go`
- 生成代码在 `internal/app/wire_gen.go`
- Provider函数在 `internal/providers/` 目录

### 消息处理流程
1. 消息从各平台（企业微信、钉钉等）进入Controller
2. Controller将消息传递给UseCase处理业务逻辑
3. UseCase调用Repository进行数据持久化
4. UseCase将处理结果返回给Controller
5. Controller将响应发送回对应平台

### 扩展新平台
1. 在 `entity/` 中定义平台特定的消息结构
2. 在 `controller/` 中实现平台适配器
3. 在 `usecase/` 中添加平台特定的业务逻辑
4. 在 `providers/` 中注册新的依赖

# 数据库设计模式

## Repository模式

### 接口设计
所有数据访问都通过Repository接口进行：

```go
type Repository interface {
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id string) (interface{}, error)
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters map[string]interface{}) ([]interface{}, error)
}
```

### GORM集成
项目使用 GORM 作为 ORM 框架：
- 模型定义在 `internal/entity/` 目录
- Repository实现在 `internal/repo/` 目录
- 数据库配置在 `config/` 目录

### 数据库迁移
使用 golang-migrate 进行数据库版本管理：
- 迁移文件存放在 `migrations/` 目录
- 文件命名格式：`000001_initial_schema.up.sql`
- 支持向上和向下迁移

## 实体设计

### 基础实体结构
为了同时获得全局唯一性和高性能的、基于时间的排序能力，主键ID采用 `UUIDv7`。

```go
type BaseEntity struct {
    ID        string         `gorm:"type:varchar(36);primaryKey"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### 消息实体
```go
type Message struct {
    BaseEntity
    PlatformID   string    `gorm:"not null;index"`
    FromUserID   string    `gorm:"not null;index"`
    ToUserID     string    `gorm:"index"`
    MessageType  string    `gorm:"not null"`
    Content      string    `gorm:"type:text"`
    Status       string    `gorm:"default:'pending'"`
    Metadata     string    `gorm:"type:json"`
    ProcessedAt  *time.Time
}
```

### 通道 (Channel) 实体
`Channel` 实体负责管理单个平台连接的配置和状态，它承担了原设计中 `Platform` 实体的职责，将平台类型和实例配置绑定在一起。

```go
type Channel struct {
    BaseEntity
    PlatformType         string           `gorm:"column:platform_type;type:varchar(50);not null"`
    ChannelName          string           `gorm:"column:channel_name;type:varchar(100);not null"`
    Config               JSONField        `gorm:"column:config;type:json"`
    AccessToken          string           `gorm:"column:access_token;type:varchar(500)"`
    AccessTokenExpiresAt *time.Time       `gorm:"column:access_token_expires_at"`
    Status               Status           `gorm:"column:status;type:tinyint;default:1"`
}
```

## 查询优化

### 索引策略
- 为经常查询的字段添加索引
- 组合索引优于多个单列索引
- 避免在小表上创建过多索引

### 分页查询
```go
type PaginationParams struct {
    Page     int `json:"page" validate:"min=1"`
    PageSize int `json:"page_size" validate:"min=1,max=100"`
}

func (r *messageRepo) ListWithPagination(ctx context.Context, params PaginationParams) (*PaginatedResult, error) {
    offset := (params.Page - 1) * params.PageSize
    // 查询逻辑
}
```

### 事务处理
```go
func (r *repository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return r.db.WithContext(ctx).Transaction(fn)
}
```

## 缓存策略

### Redis集成
- 使用Redis作为缓存层
- 缓存热点数据和查询结果
- 设置合理的过期时间

### 缓存模式
- **缓存穿透**: 使用布隆过滤器
- **缓存雪崩**: 设置随机过期时间
- **缓存击穿**: 使用分布式锁

## 监控和性能

### 查询监控
- 记录慢查询
- 监控数据库连接数
- 追踪查询性能指标

### 连接池配置
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```
