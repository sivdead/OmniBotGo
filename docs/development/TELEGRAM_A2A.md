# Telegram ↔ A2A 垂直切面（私有 MVP）

OmniBotGo 作为 **薄 IM 网关 + A2A client（C3）**：不内置 LLM。入站 Telegram 文本转发到外部 A2A Agent，再把最终文本回覆到同一 chat。

## 为何用独立 `cmd/tg-a2a`

全量 `cmd/app` 路径依赖 DB 通道、`ConnectionManager`、路由/处理器。此前出站常因 `channel.connection_status` 未随 long-poll `Start` 写回库而报 **「通道未就绪」**。

本垂直切面选择 **最小可靠金路径**：

| 组件 | 作用 |
|------|------|
| `internal/a2a` | 拉取 Agent Card + `message/send`，解析 Task/Message 最终文本 |
| `cmd/tg-a2a` | 复用 `internal/adapter/telegram` long-poll；handler 内调 A2A 再 `SendMessage` |
| 旁路 | 不经 DB / ConnectionManager / processor |

同时已修复全量路径：`ConnectionManager` 在 Stream `Start`/`Stop` 时持久化 `connection_status`；`SendMessage` 仅要求通道 **激活**（未连接只告警仍尝试发送）。

## 配置

优先级：环境变量 > `config.yaml` > 默认值。

| 项 | 环境变量 | YAML | 默认 |
|----|----------|------|------|
| A2A 基址 | `A2A_BASE_URL` | `a2a.base_url` | `http://127.0.0.1:10000` |
| Bot Token | `TELEGRAM_BOT_TOKEN` | — | 或文件 `TELEGRAM_BOT_TOKEN_FILE` / `/home/box/.config/omnibotgo/telegram_bot_token` |

```yaml
a2a:
  base_url: "http://127.0.0.1:10000"
```

外部 Agent 需提供 `GET /.well-known/agent.json`（A2A 0.3.0 JSON-RPC）。样例：LangGraph Currency Agent。

**切勿**把 bot token 或含密钥的 `.env` 提交进仓库。

## 运行

```bash
# 单元测试
go test ./internal/a2a/...

# A2A 冒烟（不需要 Telegram）
./scripts/smoke_a2a.sh

# 金路径（需要 token + 外部 Agent 在 :10000）
export A2A_BASE_URL=http://127.0.0.1:10000
go run ./cmd/tg-a2a
# 然后在 Telegram 向 bot 发：How much is 10 USD to EUR?
```

## 预期行为

1. 启动时尝试拉取 Agent Card 并打日志。
2. 收到用户文本 → `message/send`（blocking）→ 从 Task `artifacts`（或 history / Message）取出文本。
3. 原 chat `sendMessage` 回覆；A2A 失败时回「Sorry, the agent is unavailable...」。
