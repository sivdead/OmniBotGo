# OmniBotGo Phase 3 实施计划 (修订版)

## 📋 概述

本文档详细说明OmniBotGo项目Phase 3阶段的实施计划，包括各个功能模块的具体实施步骤、技术要点和预期成果。本计划遵循**"演进式设计"**和**"实用主义"**原则，旨在避免过度设计，确保快速交付高价值功能。

## 🗓️ 总体时间规划 (修订版)

- **总周期**: 8周
- **Sprint划分**: 4个Sprint (每个Sprint 2周)

## 📊 Sprint规划详情 (修订版)

### Sprint 1: 核心消息处理 (W1-2)

#### W1: 消息队列系统实现
1.  **队列模型设计**: 审查`internal/entity/queue.go`，确认字段满足需求（暂时**移除`Priority`**）。更新数据库迁移脚本。
2.  **入队与调度逻辑**: 在`usecase/queue.go`中实现消息入队，以及基于数据库的轮询调度逻辑。
3.  **失败与重试**: 实现消息处理失败后的重试计数和最大重试次数后的失败标记。
4.  **监控API**: 在`controller/http/v1/queue.go`中添加获取队列状态（如待处理、失败数量）的API。

#### W2: 消息去重机制
1.  **引入Redis**: 在`internal/infra`或类似目录中添加Redis客户端的初始化和连接管理。通过DI注入到`MessageUseCase`。
2.  **消息指纹**: 在`usecase/service/message_service.go`中实现基于`platform_message_id`等关键信息的消息指纹生成算法。
3.  **去重逻辑**: 在`usecase/message.go`的消息处理入口，使用Redis的`SETNX`或类似命令实现基于时间窗口的去重检查。

### Sprint 2: 系统扩展与平台适配 (W3-4)

#### W3: 消息处理中间件
1.  **定义接口**: 在`usecase/contracts.go`中定义`MessageHandler`接口和`MessageMiddleware`函数类型。
2.  **实现调用链**: 在`MessageUseCase`中添加一个`[]MessageMiddleware`字段，并实现中间件的链式调用逻辑。
3.  **开发示例**: 在`usecase/middleware/`下创建具体的中间件文件，如`logging.go`, `filtering.go`。

#### W4: 飞书适配器完善
1.  **事件处理**: 在`adapter/feishu/adapter.go`中，解析并处理飞书平台推送的各类事件消息。
2.  **富文本支持**: 根据飞书的API文档，扩展`adapter/feishu/types.go`以支持交互式卡片等富文本消息的发送和接收。

### Sprint 3: AI功能增强 (W5-6)

#### W5: Claude & Gemini 模型支持
1.  **Claude客户端**: 创建`internal/processor/ai/claude.go`，实现与Anthropic Claude API的对接。
2.  **Gemini客户端**: 创建`internal/processor/ai/gemini.go`，实现与Google Gemini API的对接。
3.  **统一配置**: 确保`config`模块可以灵活配置不同AI供应商的API Key和模型参数。

#### W6: AI工具调用与流式处理优化
1.  **工具扩展**: 在`internal/processor/ai/tools/`目录下，增加更多内置工具的实现。
2.  **流式处理**: 审查`internal/processor/ai/streaming.go`，优化对不同模型流式响应的处理，确保数据块解析的健壮性。

### Sprint 4: 运维与高级功能探索 (W7-8)

#### W7: Prometheus监控与日志优化
1.  **Prometheus集成**: 在`internal/metrics/`下定义核心业务指标（如消息收发数、队列大小、API延迟），并暴露`/metrics`端点。
2.  **结构化日志**: 全面审查`pkg/logger/`和项目中的日志记录点，确保输出的是带上下文的结构化日志。

#### W8: 探索部分配置热加载
1.  **限定范围**: 明确只针对无状态、变动频繁的配置进行热加载尝试，例如消息路由规则。
2.  **实现**: 增强`service/config_watcher.go`，当监听到特定配置文件变更时，通过channel通知相关`UseCase`更新其内部状态。

## 🔧 技术实现详情 (修订版)

### 1. 消息中间件
放弃复杂的插件生命周期管理，采用标准的Go Web框架中间件模式。
```go
// file: internal/usecase/contracts.go
type MessageHandler interface {
    Handle(ctx context.Context, msg *entity.Message) (*entity.Message, error)
}
type MessageMiddleware func(next MessageHandler) MessageHandler

// file: internal/usecase/message.go
func (uc *MessageUseCase) processInboundMessageWithMiddleware(ctx context.Context, msg *entity.Message) error {
    handler := uc.coreMessageHandler // coreMessageHandler是核心业务逻辑
    // 从后向前应用中间件
    for i := len(uc.middlewares) - 1; i >= 0; i-- {
        handler = uc.middlewares[i](handler)
    }
    _, err := handler.Handle(ctx, msg)
    return err
}
```

### 2. Redis用于去重
明确Redis的首要应用场景是解决具体问题，而非构建通用缓存层。
```go
// file: internal/usecase/message.go
func (uc *MessageUseCase) isDuplicate(ctx context.Context, message *entity.Message) (bool, error) {
    fingerprint := uc.messageService.GenerateMessageFingerprint(message)
    key := "dedup:" + fingerprint
    
    // SET key value EX 600 NX
    // 如果键不存在，则设置并返回true；如果键已存在，则返回false。
    // 这一个原子操作就完成了检查和设置。
    wasSet, err := uc.redisClient.SetNX(ctx, key, 1, 10*time.Minute).Result()
    if err != nil {
        uc.logger.Error("Redis check duplicate failed", "error", err)
        return false, nil // Redis故障，选择放行消息
    }
    
    // 如果wasSet为true，说明是新消息；如果为false，说明是重复消息。
    isDuplicate := !wasSet
    return isDuplicate, nil
}
```

### 3. 配置热加载（简化版）
接受重启是大部分配置变更的标准操作。仅对极少数安全且必要场景做热加载。
```go
// file: internal/service/config_watcher.go
func (cw *ConfigWatcher) watchRoutingRules() {
    // 使用 fsnotify 或类似库监听路由规则文件的变化
    // ...
    // 当检测到变化时:
    newRules, err := loadRoutingRules()
    if err == nil {
        // 通过channel将新规则发送给MessageUseCase
        cw.routingUpdateChannel <- newRules 
    }
}

// file: internal/usecase/message.go
func (uc *MessageUseCase) listenForRoutingUpdates() {
    go func() {
        for newRules := range uc.routingUpdateChannel {
            uc.mu.Lock()
            uc.routingRules = newRules // 安全地更新路由规则
            uc.mu.Unlock()
            uc.logger.Info("Routing rules reloaded dynamically.")
        }
    }()
}
```

## 📌 风险评估与缓解措施 (修订版)

| 风险 | 影响 | 可能性 | 缓解措施 |
|------|------|--------|----------|
| 消息队列性能瓶颈 (数据库) | 中 | 中 | 初期业务量不大，可接受。后期若成为瓶颈，可平滑迁移至Redis List或专业MQ。设计上保持接口化。 |
| Redis单点故障 | 高 | 低 | 消息去重失败时，日志记录并放行消息，保证主流程可用。关键数据仍以数据库为准。 |
| 中间件逻辑混乱 | 低 | 中 | 保持中间件职责单一，通过DI注入依赖而非全局状态，编写清晰的单元测试。 |

## 🔍 验收标准

1. **功能验收**
   - 所有计划功能按照设计规范实现
   - 通过所有单元测试和集成测试
   - 满足所有API规范要求

2. **性能验收**
   - 消息处理延迟<100ms (P95)
   - 队列处理吞吐量>100条/秒
   - API响应时间<200ms (P95)

3. **可靠性验收**
   - 系统可用性>99.9%
   - 消息丢失率<0.001%
   - 成功处理所有边缘情况

4. **安全验收**
   - 通过安全代码审查
   - 所有敏感信息加密存储
   - API访问控制完善

## 📚 参考资料

- [消息队列最佳实践](https://example.com/message-queue-best-practices)
- [Go插件系统设计模式](https://example.com/go-plugin-patterns)
- [Redis缓存策略](https://example.com/redis-caching-strategies)
- [Claude API文档](https://docs.anthropic.com/claude/reference/getting-started-with-the-api)
- [Gemini API文档](https://ai.google.dev/docs/gemini_api_overview)
- [Prometheus监控指南](https://example.com/prometheus-monitoring-guide) 