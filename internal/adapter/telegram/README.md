# Telegram 适配器（MVP long-poll）

私有 MVP 金路径：通过 Bot API **long-poll `getUpdates`** 收消息，用 **`sendMessage`** 回文本。不依赖第三方 SDK（stdlib `net/http`）。

## 配置

通道 `platform_type` = `telegram`，`config` 必填：

| 字段 | 必填 | 说明 |
|------|------|------|
| `bot_token` | ✅ | 向 [@BotFather](https://t.me/BotFather) 申请的 Bot Token |

`channel_id` 由 ConnectionManager 启动 Stream 时自动注入，无需手写。

示例：

```json
{
  "bot_token": "123456:ABC-DEF..."
}
```

## 工作原理

1. `ConnectionManager` 加载活跃通道；若适配器实现 `StreamAdapter`，调用 `Start`。
2. 适配器后台循环：`GET /bot<token>/getUpdates?timeout=30&offset=...`
3. 解析用户文本消息为 `dto.UnifiedMessage`（跳过 bot / 空文本）。
4. 出站：`POST /bot<token>/sendMessage`，`chat_id` = `message.ReceiverID`（或 `RawContent.chat_id`）。

> MVP **不支持 webhook 模式**；单实例建议对应一个 Telegram bot 通道。

## 创建通道（操作步骤）

1. 在 Telegram 找 `@BotFather` → `/newbot` → 得到 `bot_token`。
2. （可选）先私聊一次你的 bot，便于拿到 chat id；适配器也会从入站消息里写入 `chat_id`。
3. 在 OmniBotGo 创建通道，例如：

```http
POST /api/v1/channels
Content-Type: application/json

{
  "bot_id": "<已有 bot uuid>",
  "name": "telegram-mvp",
  "platform_type": "telegram",
  "config": {
    "bot_token": "<YOUR_BOT_TOKEN>"
  },
  "status": 1
}
```

4. 重启服务或等待 ConnectionManager 拉起 long-poll；向 bot 发文本应进入统一消息管线并可回覆。

## 接口能力

| MessageSender | WebhookProcessor | TokenManager | StreamAdapter |
|:-------------:|:----------------:|:------------:|:-------------:|
| ✅ 文本 | ❌ | ❌ | ✅ long-poll |

## 测试

```bash
go test ./internal/adapter/telegram/...
```
