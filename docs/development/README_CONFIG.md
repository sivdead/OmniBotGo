# OmniBotGo 配置系统

本项目使用 [Viper](https://github.com/spf13/viper) 作为配置管理库，它是 Go 中最流行和功能强大的配置管理解决方案。

## 项目结构

```
OmniBotGo/
├── internal/config/          # 配置管理代码
│   └── config.go            # Viper 配置实现
├── config/                  # 配置文件目录
│   ├── config.yaml         # 主配置文件
│   └── config.local.yaml   # 本地开发配置（可选）
└── README_CONFIG.md         # 配置系统文档
```

## Viper 的优势

### 多格式支持
- ✅ **YAML** (主要使用)
- ✅ **JSON** 
- ✅ **TOML**
- ✅ **HCL**
- ✅ **环境变量**
- ✅ **命令行参数**

### 配置来源优先级

Viper 按以下优先级加载配置（高优先级覆盖低优先级）：

1. **显式调用 Set 函数**
2. **环境变量** (如 `DB_DSN`, `HTTP_PORT`)
3. **配置文件** (`config.yaml`)
4. **默认值**

### 环境变量映射

配置键会自动映射到环境变量，使用下划线替换点号：

| 配置键 | 环境变量 |
|--------|----------|
| `app.name` | `APP_NAME` |
| `db.dsn` | `DB_DSN` |
| `http.port` | `HTTP_PORT` |
| `log.level` | `LOG_LEVEL` |

## 配置文件

### 主配置文件: `config/config.yaml`

```yaml
# 应用程序配置
app:
  name: "OmniBotGo"
  version: "1.0.0"

# HTTP 服务器配置
http:
  port: "8080"
  use_prefork_mode: false

# 日志配置
log:
  level: "debug"

# 数据库配置
db:
  type: "mysql"
  dsn: "user:password@tcp(localhost:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local"
  max_connections: 10

# gRPC 服务器配置
grpc:
  port: "8081"

# RabbitMQ 配置
rmq:
  server_exchange: "rpc_server"
  client_exchange: "rpc_client"
  url: "amqp://guest:guest@localhost:5672/"

# 监控配置
metrics:
  enabled: true

# Swagger 文档配置
swagger:
  enabled: true
```

## 运行模式

### 1. 标准模式
```bash
just run
```
使用配置文件的默认值

### 2. 开发模式
```bash
just run-dev
```
自动启用调试日志和 Swagger 文档

### 3. 生产模式
```bash
just run-prod
```
使用生产优化设置（无数据库迁移）

### 4. 环境变量覆盖
```bash
# 临时覆盖日志级别
LOG_LEVEL=error just run

# 覆盖数据库连接
DB_DSN="user:newpass@tcp(localhost:3306)/testdb" just run
```

## 配置验证

系统会自动验证必需的配置项：

- ✅ `app.name` - 应用程序名称
- ✅ `app.version` - 版本号
- ✅ `http.port` - HTTP端口
- ✅ `log.level` - 日志级别
- ✅ `db.dsn` - 数据库连接字符串
- ✅ `grpc.port` - gRPC端口
- ✅ `rmq.*` - RabbitMQ配置

## 默认值

如果配置文件不存在或某些值未设置，系统会使用以下默认值：

```yaml
app:
  name: "OmniBotGo"
  version: "1.0.0"
http:
  port: "8080"
  use_prefork_mode: false
log:
  level: "info"
db:
  type: "mysql"
  max_connections: 10
grpc:
  port: "8081"
rmq:
  server_exchange: "rpc_server"
  client_exchange: "rpc_client"
metrics:
  enabled: true
swagger:
  enabled: false
```

## 配置文件查找路径

Viper 会在以下路径查找配置文件：

1. `./config.yaml` (当前目录) 
2. `./config/config.yaml` (config子目录) - **推荐位置**

## 迁移和全局访问

项目还提供了全局配置访问函数，用于迁移脚本等场景：

```go
import "github.com/sivdead/OmniBotGo/internal/config"

// 获取字符串配置
dsn := config.GetString("db.dsn")

// 获取整数配置
maxConn := config.GetInt("db.max_connections")

// 获取布尔配置
swaggerEnabled := config.GetBool("swagger.enabled")
```

## 架构优势

### 清晰的职责分离

- **`internal/config/`**: 配置管理逻辑，类型定义，验证规则
- **`config/`**: 配置文件存储，环境特定的配置
- **全局访问**: 通过 `config.GetString()` 等函数提供便利的全局访问

### 符合 Go 最佳实践

1. **内部包**: 配置逻辑在 `internal/` 下，防止外部导入
2. **单一职责**: 配置代码和配置文件分离
3. **依赖注入**: 通过 Wire 注入配置，便于测试
4. **类型安全**: 强类型配置结构体，编译时检查

## 安全注意事项

⚠️ **重要安全提示**：
- **不要将包含敏感信息的 `config.yaml` 提交到版本控制系统**
- 生产环境配置应该通过环境变量或安全的配置管理系统提供
- 确保配置文件的文件权限设置正确（建议 600 权限）
- 使用 `.gitignore` 排除所有包含实际配置的文件
- 密钥、Token等敏感信息必须加密存储或使用密钥管理服务

### 最佳实践
```bash
# 设置配置文件权限
chmod 600 config/config.yaml

# 使用环境变量传递敏感信息
export DB_PASSWORD="your_secure_password"
export WECOM_SECRET="your_secret_key"

# 或使用密钥管理服务（如 Vault、AWS Secrets Manager等）
```

## 故障排除

1. **配置文件未找到**: 系统会显示警告并使用环境变量和默认值
2. **必需配置缺失**: 启动时会显示详细的验证错误
3. **环境变量格式**: 确保环境变量使用正确的命名格式（大写，下划线分隔）
4. **导入路径**: 确保使用 `github.com/sivdead/OmniBotGo/internal/config` 导入配置包
5. **权限问题**: 确保应用有权限读取配置文件 