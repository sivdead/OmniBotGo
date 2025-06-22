# 钉钉企业应用适配器

钉钉企业应用适配器实现了与钉钉企业内部应用的集成，支持工作通知、群消息、事件回调等功能。

## 功能特性

### 支持的消息类型
- **发送消息**
  - 文本消息 (text)
  - Markdown消息 (markdown)
  - 卡片消息 (action_card)
  - 链接消息 (link)
  - 图片消息 (image) - 需要先上传获取media_id
  - 文件消息 (file) - 需要先上传获取media_id
  - 图文消息 (news)

- **接收事件**
  - 组织架构事件（用户入职/离职、部门变更等）
  - 应用事件（用户激活/停用应用）
  - 审批事件（审批任务/实例变更）
  - 考勤事件（打卡、考勤结果）
  - 日程事件

### 消息发送模式
1. **工作通知** - 通过应用向用户发送消息
2. **群消息** - 向群聊发送消息
3. **DING消息** - 发送紧急DING消息

### 主要功能
- ✅ 获取和管理Access Token
- ✅ 发送工作通知
- ✅ 发送群消息
- ✅ 处理Webhook事件回调
- ✅ 支持多种消息类型
- ✅ 事件签名验证

## 使用方法

### 1. 配置参数

```go
config := map[string]interface{}{
    "app_key":       "your_app_key",       // 必填：应用AppKey
    "app_secret":    "your_app_secret",    // 必填：应用AppSecret
    "agent_id":      "your_agent_id",      // 必填：应用AgentID
    "corp_id":       "your_corp_id",       // 可选：企业ID
    "message_mode":  "work_notice",        // 可选：默认消息模式
    "webhook_token": "your_webhook_token", // 可选：Webhook验证Token
    "aes_key":       "your_aes_key",       // 可选：消息加密密钥
}
```

### 2. 通过AdapterManager使用

```go
// 获取钉钉企业应用适配器
adapter, err := adapterManager.GetAdapter(entity.PlatformTypeDingtalk)
if err != nil {
    log.Fatal("Failed to get dingtalk adapter:", err)
}

// 获取Access Token
tokenResp, err := adapter.(port.TokenManager).GetAccessToken(ctx, config)
if err != nil {
    log.Fatal("Failed to get access token:", err)
}

// 发送消息
message := &dto.UnifiedMessage{
    MessageType:  entity.MessageTypeText,
    Content:      "这是一条工作通知",
    ReceiverID:   "user123", // 用户ID
    ReceiverType: entity.ReceiverTypeUser,
}

err = adapter.(port.MessageSender).SendMessage(ctx, message, config, tokenResp.Token)
if err != nil {
    log.Error("Failed to send message:", err)
}
```

### 3. 发送不同类型的消息

#### 发送Markdown消息
```go
message := &dto.UnifiedMessage{
    MessageType: entity.MessageTypeMarkdown,
    MarkdownContent: &entity.MarkdownMessage{
        Title:   "项目进度更新",
        Content: "## 本周完成\n- 任务1\n- 任务2\n\n## 下周计划\n- 任务3",
    },
    ReceiverID:   "user123",
    ReceiverType: entity.ReceiverTypeUser,
}
```

#### 发送卡片消息
```go
message := &dto.UnifiedMessage{
    MessageType: entity.MessageTypeCard,
    CardContent: &entity.CardMessage{
        Title:   "审批提醒",
        Content: "您有一条待审批的申请",
        Buttons: []entity.CardButton{
            {
                Title:     "查看详情",
                ActionURL: "https://example.com/approval/123",
            },
        },
    },
    ReceiverID:   "user123",
    ReceiverType: entity.ReceiverTypeUser,
}
```

#### 发送群消息
```go
message := &dto.UnifiedMessage{
    MessageType:  entity.MessageTypeText,
    Content:      "群公告：明天下午2点开会",
    ReceiverID:   "chat123", // 群ID
    ReceiverType: entity.ReceiverTypeGroup,
}
```

### 4. 处理Webhook事件

```go
// 在HTTP处理器中
func handleDingtalkWebhook(w http.ResponseWriter, r *http.Request) {
    // 处理Webhook消息
    unifiedMsg, err := adapter.(port.WebhookProcessor).ProcessWebhookMessage(ctx, r, channelID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 处理消息
    err = messageUsecase.ProcessInboundMessage(ctx, unifiedMsg)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

## 获取配置参数

1. 登录[钉钉开放平台](https://open.dingtalk.com)
2. 创建企业内部应用
3. 获取应用的AppKey、AppSecret和AgentID
4. 配置事件订阅和消息接收

## 注意事项

1. **Access Token管理**：Token有效期为2小时，适配器会自动缓存和刷新
2. **消息发送限制**：注意钉钉的API调用频率限制
3. **文件上传**：发送图片和文件需要先调用上传接口获取media_id
4. **签名验证**：生产环境建议启用Webhook签名验证
5. **权限配置**：确保应用有相应的权限（如发送工作通知、访问通讯录等）

## 扩展功能

如需添加更多功能，可以扩展适配器：

1. **媒体文件上传**：添加上传图片、文件等功能
2. **用户信息获取**：添加获取用户详情、部门信息等功能
3. **更多消息类型**：支持OA消息、待办消息等
4. **批量操作**：支持批量发送消息、批量获取用户信息等 