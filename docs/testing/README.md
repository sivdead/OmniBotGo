# OmniBotGo 测试文档

本目录包含了OmniBotGo项目的完整测试文档，涵盖了测试策略、实施指南、模板和配置。

## 📚 文档导航

### 核心文档

#### [测试用例清单](测试用例清单.md)
- 完整的测试用例清单，基于项目设计文档和整洁架构原则
- 包含单元测试、集成测试、架构测试和性能测试
- 测试覆盖率目标和维护策略

#### [测试执行指南](测试执行指南.md)
- 具体的测试执行方法和最佳实践
- 单元测试、集成测试、性能测试的实施步骤
- 测试工具使用和调试技巧

#### [测试配置指南](测试配置指南.md)
- 完整的测试环境配置说明
- CI/CD集成配置
- 本地开发环境设置
- 故障排除和调试技巧

#### [测试模板](测试模板.md)
- 各种测试场景的代码模板
- Entity、UseCase、Repository、Adapter、Controller层测试模板
- 集成测试和性能测试模板
- 测试工具函数和夹具模板

## 🧪 测试架构概览

### 测试分层策略
```
单元测试 (Unit Tests)
├── Entity层测试 - 业务实体和领域模型
├── UseCase层测试 - 核心业务逻辑
├── Repository层测试 - 数据访问层
├── Adapter层测试 - 平台适配器
└── Controller层测试 - HTTP/gRPC控制器

集成测试 (Integration Tests)
├── API集成测试 - 端到端API测试
├── 数据库集成测试 - 数据持久化测试
├── 平台集成测试 - 外部平台对接测试
└── 消息流程集成测试 - 完整消息处理流程

架构测试 (Architecture Tests)
├── 依赖关系测试 - 整洁架构依赖规则
├── 接口契约测试 - Port接口实现验证
└── 组件隔离测试 - 模块边界验证

性能测试 (Performance Tests)
├── 负载测试 - 高并发消息处理
├── 压力测试 - 系统极限测试
└── 基准测试 - 关键组件性能
```

## 🚀 快速开始

### 环境准备
```bash
# 1. 安装测试依赖
go mod download

# 2. 生成Mock文件
just mock

# 3. 生成Wire代码
just wire

# 4. 运行所有测试
just test
```

### 常用测试命令
```bash
# 单元测试
go test ./internal/...

# 集成测试
go test ./integration-test/...

# 架构测试
go test ./test/architecture/...

# 带覆盖率的测试
go test -cover ./internal/...

# 性能基准测试
go test -bench=. ./internal/...
```

## 📊 测试覆盖率目标

### 覆盖率要求
- **Entity层**: 90%+
- **UseCase层**: 85%+
- **Repository层**: 80%+
- **Adapter层**: 75%+
- **Controller层**: 70%+

### 关键路径覆盖
- **消息处理流程**: 100%
- **错误处理路径**: 95%+
- **安全验证逻辑**: 100%

## 🛠️ 测试工具栈

### 核心工具
- **testify** - 断言和测试套件
- **gomock** - Mock对象生成
- **testcontainers** - 集成测试容器
- **httptest** - HTTP服务测试

### 辅助工具
- **wire** - 依赖注入代码生成
- **golangci-lint** - 静态代码分析
- **k6** - API负载测试
- **delve** - 测试调试器

## 📁 项目测试结构

```
OmniBotGo/
├── internal/
│   ├── entity/
│   │   ├── base_test.go
│   │   ├── example_test.go
│   │   └── ...
│   ├── usecase/
│   │   ├── mocks_repo_test.go
│   │   ├── mocks_usecase_test.go
│   │   └── ...
│   ├── adapter/
│   │   ├── manager_test.go
│   │   ├── dingtalk_enterprise/
│   │   │   └── adapter_test.go
│   │   └── ...
│   └── ...
├── test/
│   ├── architecture/
│   │   └── dependencies_test.go
│   └── integration/
│       └── wire_test.go
├── integration-test/
│   └── integration_test.go
├── testdata/
│   ├── fixtures/
│   ├── golden/
│   └── mocks/
└── docs/testing/
    ├── README.md
    ├── 测试用例清单.md
    ├── 测试执行指南.md
    ├── 测试配置指南.md
    └── 测试模板.md
```

## 🔄 测试工作流程

### 1. 开发阶段
1. 编写单元测试（TDD方式）
2. 实现功能代码
3. 运行单元测试验证
4. 添加集成测试

### 2. 代码提交前
1. 运行完整测试套件
2. 检查测试覆盖率
3. 运行静态代码分析
4. 验证架构依赖规则

### 3. CI/CD流程
1. 自动运行所有测试
2. 生成测试报告
3. 上传覆盖率数据
4. 性能基准测试

## 📈 测试监控

### 持续监控指标
- 测试通过率
- 代码覆盖率
- 测试执行时间
- 失败测试分析

### 质量门禁
- 所有测试必须通过
- 覆盖率不能低于目标值
- 不能引入新的架构违规
- 性能不能显著下降

## 🎯 最佳实践

### 测试编写原则
1. **独立性** - 每个测试都应该独立运行
2. **可重复性** - 测试结果应该一致
3. **快速性** - 单元测试应该快速执行
4. **清晰性** - 测试意图应该明确

### 命名规范
- 测试函数: `Test<FunctionName>_<Scenario>`
- 基准测试: `Benchmark<FunctionName>_<Scenario>`
- 测试文件: `*_test.go`

### Mock使用原则
- 只Mock外部依赖
- 使用接口进行Mock
- 验证Mock调用
- 避免过度Mock

## 🔧 故障排除

### 常见问题
1. **Mock生成失败** - 检查mockgen版本和接口定义
2. **数据库连接失败** - 确认Docker服务状态
3. **测试超时** - 增加超时时间或检查死锁
4. **覆盖率不足** - 添加缺失的测试用例

### 调试技巧
1. 使用`-v`标志获取详细输出
2. 使用`dlv`调试器进行断点调试
3. 启用详细日志输出
4. 使用`-race`检测竞态条件

## 📝 贡献指南

### 添加新测试
1. 遵循现有的测试结构和命名规范
2. 使用提供的测试模板
3. 确保测试覆盖率达到要求
4. 添加必要的文档说明

### 更新测试文档
1. 保持文档与代码同步
2. 更新测试用例清单
3. 补充新的测试模板
4. 完善配置说明

---

通过遵循本测试文档，可以确保OmniBotGo项目的高质量和稳定性。如有疑问或建议，请参考具体的文档页面或联系开发团队。 