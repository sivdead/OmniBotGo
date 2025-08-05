# OmniBotGo 项目文档

欢迎查阅 OmniBotGo 项目文档。本项目是一个使用 Go 语言开发的高性能、可扩展的统一消息适配器网关。

## 📚 文档导航

### 🏗️ 架构设计
- [**架构概览**](architecture/README.md) - 系统架构设计和整洁架构原则
- [**技术架构文档**](architecture/02_technical_architecture.md) - 详细的技术架构设计
- [**功能架构文档**](architecture/03_functional_architecture.md) - 系统功能架构和模块划分

### 💡 设计文档
- [**架构与数据库设计**](ARCHITECTURE_DESIGN.md) - 核心架构、实体及数据库设计原则
- [**需求说明**](design/需求说明.md) - 项目功能需求和非功能需求
- [**数据库ER图**](design/数据库ER图.md) - 数据库实体关系图

### 🛠️ 开发指南
- [**配置系统说明**](development/README_CONFIG.md) - 配置管理系统详细说明
- [**开发计划**](development/开发计划.md) - 项目开发计划和里程碑
- [**实现完成总结**](development/实现完成总结.md) - 各模块实现状态和技术特点

### 🔌 API文档
- [**Swagger API文档**](swagger.yaml) - RESTful API接口文档
- 在线查看：启动服务后访问 `http://localhost:8080/swagger`

## 🚀 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis (可选)
- RabbitMQ (可选)

### 安装运行
```bash
# 克隆项目
git clone https://github.com/sivdead/OmniBotGo.git
cd OmniBotGo

# 复制配置文件
cp config.yaml.example config.yaml

# 编辑配置文件
vim config.yaml

# 运行项目
just run
```

### 项目结构
```
OmniBotGo/
├── cmd/                    # 应用程序入口
├── internal/               # 内部包（业务逻辑）
│   ├── adapter/           # 平台适配器
│   ├── controller/        # HTTP/gRPC控制器
│   ├── usecase/           # 业务用例
│   ├── entity/            # 业务实体
│   └── repo/              # 数据访问层
├── pkg/                    # 公共包（可复用）
├── docs/                   # 项目文档
├── migrations/             # 数据库迁移
└── config.yaml.example     # 配置文件模板
```

## 📊 项目状态

- **架构设计**: ✅ 完成（整洁架构 + DDD）
- **核心功能**: ✅ 完成（消息处理、路由、适配器管理）
- **平台支持**: ✅ 企业微信、钉钉、飞书
- **API接口**: ✅ 完成（RESTful + Swagger）
- **数据库层**: ✅ 完成（MySQL + GORM）
- **测试覆盖**: 🚧 进行中（目标80%+）

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request。在提交代码前，请确保：

1. 代码通过所有测试：`just test`
2. 代码符合规范：`just lint`
3. 更新相关文档
4. 提交信息遵循约定式提交规范

## 📝 许可证

本项目采用 MIT 许可证，详见 [LICENSE](../LICENSE) 文件。 