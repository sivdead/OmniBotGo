# 1. 重构方案与计划

本文档旨在阐述对 OmniBotGo 项目进行核心架构重构的详细方案和执行计划。

## 1.1. 背景与动机 (Why)

当前项目架构虽基于整洁架构思想，但在实践中暴露了两个核心问题，限制了其可扩展性和可维护性：

1.  **接口设计违反"接口隔离原则"和"依赖倒置原则"**:
    *   `internal/entity/PlatformAdapter` 接口过于庞大（"胖接口"），它将多种不同的能力（Webhook处理、API调用、Token管理）捆绑在一起。
    *   这导致了像 `dingtalk_stream` 这样只需要部分能力的适配器，被迫要去实现它根本用不到的功能，产生了大量需要返回 `error("not supported")` 的"僵尸代码"。
    *   更严重的是，接口的所有权不属于客户端（`UseCase`），而是被放在了 `entity` 包中，导致了错误的依赖关系。

2.  **缺少统一的连接生命周期管理机制**:
    *   项目没有一个标准化的方式来管理需要"主动连接"的平台（如钉钉Stream模式）。
    *   这类连接的启动、停止、重连等生命周期管理逻辑散落在应用入口 (`main.go`)，导致代码硬编码、难以扩展（例如无法同时运行两个钉钉机器人）、且管理与配置脱节。

为了从根本上解决这些问题，使项目成为一个真正健壮、可扩展的统一消息网关，本次重构势在必行。

## 1.2. 重构目标 (What)

本次重构旨在达成以下目标：

1.  **技术架构优化**:
    *   **接口隔离**: 将"胖接口"拆分为多个基于能力（Capability-based）的、小而专一的接口。
    *   **依赖倒置**: 接口的所有权回归 `UseCase` 层，实现标准的整洁架构依赖关系。
    *   **代码纯粹**: 物理移除适配器中所有用不到的"僵尸代码"。

2.  **功能架构升级**:
    *   **生命周期管理**: 引入新的 `ConnectionManager` 组件，统一管理所有"主动连接"型适配器的生命周期。
    *   **配置驱动**: 实现根据配置文件动态启动、管理和停止所有平台连接的能力。
    *   **高度可扩展**: 为未来接入更多不同类型的平台（主动/被动模式）打下坚实的基础。

## 1.3. 实施计划 (How)

本次重构将分阶段进行，以确保过程平稳可控。

### 第一阶段：接口与核心组件重构

1.  **创建 `port` 包**: 在 `internal/usecase/` 下创建 `port/` 目录，用于存放所有归属于 `UseCase` 的接口定义。
2.  **定义新接口**: 在 `port/` 包中，根据能力定义新的、细粒度的接口，如 `MessageSender`, `WebhookProcessor`, `TokenManager`, `StreamAdapter` 等。
3.  **创建 `ConnectionManager`**: 在 `internal/app/` 或 `internal/` 的新包中创建 `ConnectionManager` 结构体，并定义其 `Start()` 和 `Stop()` 方法。
4.  **修改 `MessageUseCase`**: 调整 `MessageUseCase` 的依赖，使其不再依赖旧的 `AdapterManager`，而是依赖 `port/` 中定义的具体接口。

### 第二阶段：适配器迁移与重构

1.  **重构 `AdapterManager`**: 
    *   修改 `AdapterManager`，使其职责更纯粹：仅负责适配器的注册和按类型获取。
    *   增加新的方法，通过类型断言检查适配器是否实现了某个 `port` 接口，从而返回特定能力。
2.  **迁移 `dingtalk_stream`**:
    *   修改 `DingtalkStreamAdapter`，让其实现 `port.MessageSender` 和 `port.StreamAdapter` 接口。
    *   **物理删除** `GetAccessToken`, `VerifyWebhook`, `ParseInboundMessage` 等所有不相关的方法。
3.  **迁移 `wecom` 和 `feishu`**:
    *   修改 `WecomAdapter` 和 `FeishuAdapter`，让它们根据自身能力实现 `port.MessageSender`, `port.TokenManager`, `port.WebhookProcessor` 等接口。
    *   移除旧的 `PlatformAdapter` 依赖。

### 第三阶段：应用总装与启动流程改造

1.  **更新依赖注入 (Wire)**:
    *   修改 `internal/app/wire.go`，注入新的 `ConnectionManager` 和其他重构后的组件。
    *   更新 Provider 函数，以适应新的接口和结构。
2.  **改造应用入口 (`main.go`)**:
    *   在 `main` 函数中，调用 `ConnectionManager.Start()` 来启动所有主动连接。
    *   在应用接收到退出信号时，调用 `ConnectionManager.Stop()` 来实现所有连接的优雅关闭。
3.  **清理旧代码**:
    *   确认所有适配器都已迁移完毕后，从 `internal/entity/platform.go` 中**彻底删除** `PlatformAdapter` 接口的定义。
    *   删除其他因重构而不再需要的代码。

完成以上三个阶段后，项目将拥有一个全新的、更加清晰和健壮的架构。 