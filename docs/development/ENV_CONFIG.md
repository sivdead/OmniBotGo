# 环境变量配置说明

## 概述

OmniBotGo 支持通过环境变量覆盖配置文件中的设置。这对于容器化部署（Docker/Kubernetes）特别有用。

## 配置优先级

配置按以下优先级加载（优先级从高到低）：
1. 环境变量
2. 配置文件（config.yaml）
3. 默认值

## 环境变量命名规则

- 配置项路径使用下划线（_）分隔
- 所有字母大写
- 例如：配置文件中的 `app.name` 对应环境变量 `APP_NAME`

## 支持的环境变量

### 应用配置
```bash
APP_NAME=OmniBotGo              # 应用名称
APP_VERSION=1.0.0               # 应用版本
```

### HTTP服务配置
```bash
HTTP_PORT=8080                  # HTTP服务端口
HTTP_USE_PREFORK_MODE=false     # 是否使用prefork模式
```

### 日志配置
```bash
LOG_LEVEL=info                  # 日志级别（debug/info/warn/error）
```

### 数据库配置
```bash
DB_TYPE=mysql                   # 数据库类型（mysql/postgres/sqlite）
DB_DSN=user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
DB_MAX_CONNECTIONS=10           # 最大连接数
DB_LOG_LEVEL=warn              # 数据库日志级别
DB_SLOW_THRESHOLD=200          # 慢查询阈值（毫秒）
```

### gRPC配置
```bash
GRPC_PORT=8081                  # gRPC服务端口
```

### RabbitMQ配置
```bash
RMQ_URL=amqp://guest:guest@localhost:5672/
RMQ_SERVER_EXCHANGE=rpc_server  # 服务端交换机
RMQ_CLIENT_EXCHANGE=rpc_client  # 客户端交换机
```

### 监控配置
```bash
METRICS_ENABLED=true            # 是否启用监控指标
```

### Swagger配置
```bash
SWAGGER_ENABLED=false           # 是否启用Swagger文档
```

## 使用示例

### 直接运行
```bash
# 设置环境变量并运行
export DB_DSN="user:password@tcp(localhost:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local"
export HTTP_PORT=8888
export LOG_LEVEL=debug
./omnibotgo
```

### Docker运行
```bash
docker run -d \
  -e DB_DSN="user:password@tcp(mysql:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local" \
  -e HTTP_PORT=8080 \
  -e LOG_LEVEL=info \
  -p 8080:8080 \
  omnibotgo:latest
```

### Docker Compose
```yaml
version: '3.8'
services:
  omnibotgo:
    image: omnibotgo:latest
    environment:
      - DB_TYPE=mysql
      - DB_DSN=root:password@tcp(mysql:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local
      - HTTP_PORT=8080
      - LOG_LEVEL=info
      - METRICS_ENABLED=true
    ports:
      - "8080:8080"
    depends_on:
      - mysql
```

### Kubernetes ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: omnibotgo-config
data:
  APP_NAME: "OmniBotGo"
  HTTP_PORT: "8080"
  LOG_LEVEL: "info"
  DB_TYPE: "mysql"
  METRICS_ENABLED: "true"
---
apiVersion: v1
kind: Secret
metadata:
  name: omnibotgo-secret
type: Opaque
data:
  DB_DSN: <base64-encoded-dsn>
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: omnibotgo
spec:
  template:
    spec:
      containers:
      - name: omnibotgo
        image: omnibotgo:latest
        envFrom:
        - configMapRef:
            name: omnibotgo-config
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: omnibotgo-secret
              key: DB_DSN
```

## 平台特定配置

平台配置通常存储在数据库中，但也可以通过环境变量设置初始值：

### 企业微信
```bash
WECOM_CORP_ID=your_corp_id
WECOM_AGENT_ID=your_agent_id
WECOM_SECRET=your_secret
WECOM_TOKEN=your_token
WECOM_ENCODING_AES_KEY=your_encoding_aes_key
```

### 钉钉
```bash
DINGTALK_CLIENT_ID=your_client_id
DINGTALK_CLIENT_SECRET=your_client_secret
```

### 飞书
```bash
FEISHU_APP_ID=your_app_id
FEISHU_APP_SECRET=your_app_secret
```

## 最佳实践

1. **敏感信息管理**：
   - 使用环境变量存储敏感信息（如数据库密码、API密钥）
   - 在生产环境中使用密钥管理服务（如 Kubernetes Secrets、AWS Secrets Manager）

2. **配置文件与环境变量结合**：
   - 在配置文件中设置默认值和非敏感配置
   - 使用环境变量覆盖特定环境的设置

3. **环境隔离**：
   - 为不同环境（开发、测试、生产）使用不同的环境变量集
   - 使用 `.env` 文件管理本地开发环境变量（不要提交到版本控制）

4. **配置验证**：
   - 应用启动时会验证必需的配置项
   - 缺少必需配置时，应用会拒绝启动并显示错误信息

## 故障排查

1. **查看当前配置**：
   应用启动时会打印使用的配置文件路径

2. **环境变量未生效**：
   - 确保环境变量名称正确（大写，下划线分隔）
   - 检查是否有拼写错误
   - 使用 `env | grep -i app` 查看相关环境变量

3. **配置优先级问题**：
   - 环境变量总是优先于配置文件
   - 如果设置了环境变量，配置文件中的相应值会被忽略 