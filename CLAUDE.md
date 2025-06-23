# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OmniBotGo is a high-performance, scalable unified messaging adapter gateway built with Go. It enables a single instance to simultaneously connect and manage multiple instant messaging platforms (Enterprise WeChat, DingTalk, WeChat Official Account, Feishu) and relay messages to backend business logic or AI services.

## Key Development Commands

### Core Commands (via justfile)
- `just run` - Run the application (includes dependency injection generation and database migration)
- `just run-dev` - Run in development mode with debug logging and Swagger enabled
- `just run-prod` - Run in production mode
- `just test` - Run unit tests with coverage
- `just integration-test` - Run integration tests
- `just linter-golangci` - Run code quality checks
- `just format` - Format code using gofumpt and gci
- `just wire` - Generate Wire dependency injection code
- `just swag-v1` - Generate Swagger documentation
- `just deps` - Tidy and verify Go modules

### Database Commands
- `just migrate-create name` - Create new database migration
- `just migrate-up` - Apply database migrations

### Docker Commands
- `just compose-up` - Start MySQL and RabbitMQ for development
- `just compose-up-all` - Start full stack including application and nginx
- `just compose-up-integration-test` - Run integration tests in containers
- `just compose-down` - Stop all services

### Testing
- Unit tests: `go test -v -race ./internal/...`
- Integration tests: Located in `integration-test/` directory
- Always run `just linter-golangci` before committing

## Architecture Overview

### Clean Architecture Implementation
The project follows Clean Architecture principles with dependency inversion:
- **Entity Layer**: Business entities and domain models (`internal/entity/`)
- **UseCase Layer**: Business logic and port interfaces (`internal/usecase/`)
- **Adapter Layer**: Controllers, repositories, platform adapters (`internal/controller/`, `internal/repo/`, `internal/adapter/`)
- **Infrastructure**: Database, HTTP server, messaging infrastructure (`pkg/`)

### Key Components

#### Dependency Injection (Google Wire)
- Configuration: `internal/app/wire.go`
- Generated code: `internal/app/wire_gen.go` (auto-generated, don't edit)
- Provider functions: `internal/providers/` directory
- Always run `just wire` after modifying providers

#### Platform Adapters (`internal/adapter/`)
Each platform adapter implements capability-based interfaces:
- `MessageSender` - Send messages to platforms
- `WebhookProcessor` - Handle incoming webhooks  
- `TokenManager` - Manage access tokens
- `StreamAdapter` - Handle real-time connections (e.g., DingTalk Stream)
- `ConfigValidator` - Validate platform configurations

#### Connection Management
- `ConnectionManager` (`internal/service/connection_manager.go`) manages active stream connections
- Loads configuration from database and manages lifecycle of persistent connections
- Supports graceful startup/shutdown and reconnection

#### Message Flow
1. Messages enter via Controller layer (HTTP webhooks or Stream connections)
2. Controllers forward to UseCase layer for business logic
3. UseCase calls Repository for data persistence
4. UseCase returns response to Controller
5. Controller sends response back to platform

### Database Architecture
- **Primary ORM**: GORM for model definitions, CRUD operations, relationships, transactions
- **Query Builder**: Squirrel for complex queries, dynamic SQL, analytics
- **Shared connection pool** between GORM and Squirrel
- **Migrations**: Located in `migrations/` directory

## Configuration Management

### Configuration Sources (priority order)
1. Environment variables (highest priority)
2. YAML configuration file (`config.yaml`)
3. Default values (lowest priority)

### Key Configuration Files
- `config.yaml.example` - Configuration template
- `docs/development/ENV_CONFIG.md` - Environment variable documentation
- Configuration struct: `internal/config/config.go`

### Environment Variable Examples
```bash
# Database
DB_TYPE=mysql
DB_DSN="user:pass@tcp(localhost:3306)/dbname"

# Services  
HTTP_PORT=8080
GRPC_PORT=8081
LOG_LEVEL=debug

# Platform credentials
WECOM_CORP_ID=your_corp_id
DINGTALK_CLIENT_ID=your_client_id
```

## Platform Integration

### Supported Platforms
- **Enterprise WeChat (WeCom)**: Webhook-based, supports rich message types and events
- **DingTalk**: Stream and Enterprise Application modes with real-time messaging
- **WeChat Official Account**: Webhook-based subscription/service account support
- **Feishu**: Basic messaging with official SDK integration

### Adding New Platform Adapters
1. Define platform interfaces in `internal/usecase/port/`
2. Implement adapter in `internal/adapter/platform_name/`
3. Create Provider function in `internal/providers/`
4. Add to appropriate ProviderSet
5. Run `just wire` to regenerate dependency injection code
6. Add message type mappings in `internal/entity/message.go`

## Database Operations

### GORM Usage (Preferred for simple operations)
```go
// Basic CRUD
func (r *Repository) Create(ctx context.Context, entity *Entity) error {
    return r.DB.WithContext(ctx).Create(entity).Error
}

// Relationships
func (r *Repository) GetWithRelations(ctx context.Context, id uint) (*Entity, error) {
    var entity Entity
    err := r.DB.WithContext(ctx).Preload("Relations").First(&entity, id).Error
    return &entity, err
}
```

### Squirrel Usage (For complex queries)
```go
// Dynamic queries
query := r.Builder.Select("*").From("table")
if condition {
    query = query.Where("column = ?", value)
}
sql, args, _ := query.ToSql()
```

## Testing Guidelines

### Unit Testing
- Use testify/assert for assertions
- Mock dependencies using go-mock
- Test files follow `*_test.go` naming
- Maintain >80% test coverage

### Integration Testing
- Located in `integration-test/` directory
- Use Docker containers for testing
- Run with `just compose-up-integration-test`

### Mock Generation
```bash
just mock  # Generates mocks for interfaces
```

## Development Workflow

### Code Quality
- Always run `just linter-golangci` before committing
- Use `just format` to format code consistently
- Follow Go naming conventions and clean architecture principles

### Pre-commit Checklist
1. `just wire` - Regenerate dependency injection
2. `just swag-v1` - Update API documentation  
3. `just format` - Format code
4. `just linter-golangci` - Run linters
5. `just test` - Run tests

### Environment Setup
1. `just compose-up` - Start dependencies (MySQL, RabbitMQ)
2. Copy `config.yaml.example` to `config.yaml` and configure
3. `just run` - Start application

## Important Notes

- **Never edit** `internal/app/wire_gen.go` directly - it's auto-generated
- **Always run** `just wire` after modifying Provider functions
- **Configuration priority**: Environment variables override YAML configuration
- **Database migrations**: Use `just migrate-create` for new schema changes
- **Platform adapters** implement capability-based interfaces, not monolithic interfaces
- **Connection lifecycle** is managed by ConnectionManager for persistent connections
- **Message processing** follows clean architecture flow: Controller → UseCase → Repository