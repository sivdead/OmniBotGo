# 架构文档

本目录包含了 OmniBotGo 项目的架构设计和技术文档。

## 架构概览

OmniBotGo 采用整洁架构（Clean Architecture）设计，主要特点：

- **分层设计**：Entity、UseCase、Repository、Controller等层次分明
- **依赖倒置**：高层模块不依赖低层模块，都依赖抽象
- **端口适配器**：通过接口隔离业务逻辑和外部依赖
- **可测试性**：业务逻辑独立于框架，便于测试

### 核心架构原则

1. **接口隔离原则**：基于能力的细粒度接口设计
2. **依赖倒置原则**：UseCase层定义接口，Adapter层实现
3. **单一职责原则**：每个组件只负责一个功能
4. **开闭原则**：对扩展开放，对修改关闭

## 文档列表

1. [技术架构设计](02_technical_architecture.md) - 系统的技术架构详细设计
2. [功能架构设计](03_functional_architecture.md) - 系统的功能架构和模块划分

## 平台能力矩阵

| 平台 | MessageSender | WebhookProcessor | TokenManager | StreamAdapter | 连接模式 |
|------|:-------------:|:----------------:|:------------:|:-------------:|:--------:|
| **企业微信** | ✅ | ✅ | ✅ | ❌ | Webhook |
| **飞书** | ✅ | ✅ | ✅ | ❌ | Webhook |
| **钉钉Stream** | ✅ | ❌ | ❌ | ✅ | Stream |

## 快速导航

- 项目的整体设计请查看 [技术架构设计](02_technical_architecture.md)
- 系统功能和模块划分请查看 [功能架构设计](03_functional_architecture.md) 