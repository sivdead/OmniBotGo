# 导入环境变量
set dotenv-load := true

# 设置 Windows 下的 shell
set shell := ["pwsh.exe", "-c"]

# 变量定义
local_bin := justfile_directory() + "/bin"
base_stack := "docker compose -f docker/docker-compose.yml"
integration_test_stack := base_stack + " -f docker/docker-compose-integration-test.yml"
all_stack := integration_test_stack

# 默认命令 - 显示帮助
default:
    @just --list

# 运行 docker compose (不包含后端和反向代理)
compose-up:
    cd docker && docker compose up --build -d db rabbitmq && docker compose logs -f

# 运行 docker compose (包含后端和反向代理)
compose-up-all:
    cd docker && docker compose up --build -d

# 运行 docker compose 并进行集成测试
compose-up-integration-test:
    cd docker && docker compose -f docker-compose.yml -f docker-compose-integration-test.yml up --build --abort-on-container-exit --exit-code-from integration-test

# 停止 docker compose
compose-down:
    cd docker && docker compose -f docker-compose.yml -f docker-compose-integration-test.yml down --remove-orphans

# 生成 swagger 文档
swag-v1:
    swag init -g internal/controller/http/router.go

# 从 proto 文件生成源代码 (已禁用 - proto文件已删除)
proto-v1:
    @echo "Proto generation disabled - proto files removed"

# 整理和验证依赖
deps:
    go mod tidy && go mod verify

# 检查依赖漏洞
deps-audit:
    govulncheck ./...

# 运行代码格式化
format:
    gofumpt -l -w .
    gci write . --skip-generated -s standard -s default

# 运行应用 (依赖于 deps, swag-v1, proto-v1, wire)
# 配置现在从 config.yaml 文件读取，环境变量可以覆盖配置文件中的值
run: deps swag-v1 proto-v1 wire
    go mod download
    $env:CGO_ENABLED="0"; go run -tags migrate ./cmd/app

# 运行应用（开发模式，覆盖一些配置为开发友好值）
run-dev: deps swag-v1 proto-v1 wire
    go mod download
    $env:LOG_LEVEL="debug"; $env:SWAGGER_ENABLED="true"; $env:CGO_ENABLED="0"; go run -tags migrate ./cmd/app

# 运行应用（生产模式）
run-prod: deps swag-v1 proto-v1 wire
    go mod download
    $env:LOG_LEVEL="info"; $env:SWAGGER_ENABLED="false"; $env:CGO_ENABLED="0"; go run ./cmd/app

# 删除 docker volume
docker-rm-volume:
    docker volume rm go-clean-template_pg-data

# golangci lint 检查
linter-golangci:
    golangci-lint run

# hadolint lint 检查
linter-hadolint:
    git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint

# dotenv lint 检查
linter-dotenv:
    dotenv-linter

# 运行测试
test:
    go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/...

# 运行集成测试
integration-test:
    go clean -testcache && go test -v ./integration-test/...

# 生成 Wire 依赖注入代码
wire:
    wire ./internal/app

# 生成 mock 文件
mock:
    mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
    mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go

# 创建新的数据库迁移文件
migrate-create name:
    migrate create -ext sql -dir migrations '{{name}}'

# 执行数据库迁移
migrate-up:
    migrate -path migrations -database '$PG_URL?sslmode=disable' up

# 安装工具依赖
bin-deps:
    GOBIN={{local_bin}} go install tool

# 预提交检查 (依赖于多个任务)
pre-commit: swag-v1 proto-v1 wire mock format linter-golangci test 