# Access loop: bot → channel → processor → route → send/receive

Copyable end-to-end path for a fresh OmniBotGo instance. Auth and env details also live in [ENV_CONFIG.md](./ENV_CONFIG.md).

## Prerequisites

```bash
export BASE_URL="${BASE_URL:-http://localhost:8080}"
export HTTP_API_KEY="${HTTP_API_KEY:?set HTTP_API_KEY}"   # Admin Bearer key
export AUTH="Authorization: Bearer ${HTTP_API_KEY}"
```

- Migrations: dialect dirs `migrations/{mysql,postgres,sqlite}` selected by `DB_TYPE`.
- Wait for readiness: `GET ${BASE_URL}/ready` (not a placeholder always-200).
- Admin APIs under `/api/v1/*` require `HTTP_API_KEY`.
- Platform webhooks use `/webhook/{platform}/{channel_id}` (UUID) with platform signature/`verification_token` — **not** the admin API key.

## 1. Create bot (account)

```bash
curl -sS -X POST "$BASE_URL/api/v1/bots" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"bot_name":"demo-bot","bot_type":"general","description":"access loop","created_by":"ops"}'
# note data.id → BOT_ID
```

## 2. Create channel (platform backend)

```bash
curl -sS -X POST "$BASE_URL/api/v1/channels" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"bot_id\":\"$BOT_ID\",\"platform_type\":\"wecom\",\"channel_name\":\"demo-wecom\",\"config\":{\"corp_id\":\"ww_xxx\",\"agent_id\":\"1000001\",\"secret\":\"REPLACE_ME\"}}"
# note data.id → CHANNEL_ID, data.webhook_path → configure on the platform
```

## 3. Create processor + default route

```bash
curl -sS -X POST "$BASE_URL/api/v1/processors" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"demo-auto-reply","type":"auto_reply","config":{"reply_content":"你好，这是默认回复"},"priority":10}'
# note data.id → PROCESSOR_ID

curl -sS -X POST "$BASE_URL/api/v1/processors/$PROCESSOR_ID/rules" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"rule_name\":\"default\",\"processor_id\":\"$PROCESSOR_ID\",\"route_type\":1,\"priority\":100,\"is_fallback\":true,\"status\":1}"
```

## 4. Preview routing (no sends)

```bash
curl -sS -X POST "$BASE_URL/api/v1/routing/preview" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"channel_id\":\"$CHANNEL_ID\",\"message_type\":\"text\",\"content\":\"hi\",\"sender_id\":\"user-1\"}"
```

## 5. Proactive send vs reply-to-original

```bash
# Proactive send (explicit channel + receiver)
curl -sS -X POST "$BASE_URL/api/v1/messages/send" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"channel_id\":\"$CHANNEL_ID\",\"message_type\":\"text\",\"receiver_id\":\"user-1\",\"receiver_type\":\"user\",\"content\":\"hello from ops\"}"

# Reply to an inbound message (DM/group targeting resolved server-side)
curl -sS -X POST "$BASE_URL/api/v1/messages/$INBOUND_MESSAGE_ID/reply" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"content":"reply from ops","message_type":"text"}'
```

Inbound receive is via platform webhook/stream → `ProcessInboundMessage`.

## Troubleshooting

| Symptom | Check |
|--------|--------|
| Admin 401/503 | `HTTP_API_KEY` set; header `Authorization: Bearer …` |
| `/ready` 503 | DB up; migrations applied for `DB_TYPE` |
| Webhook 404 | Use `/webhook/{platform}/{channel_uuid}` from channel create response |
| Feishu/WeCom verify fail | Platform `verification_token` / signature in channel config |
| Reply goes to wrong target | Use `/messages/{id}/reply` (not blind swap); group inbound keeps group receiver |
| Processor `api_key` missing in GET | Expected — responses redact secrets (T07) |
| Message stuck failed | `POST /api/v1/messages/{message_id}/retry`; statuses: receive→process→deliver (see `MessageStatus` in code) |

Runnable script: [`scripts/init_access_loop.sh`](../../scripts/init_access_loop.sh).

Real-platform E2E (when credentials exist): [`PLATFORM_E2E_CHECKLIST.md`](./PLATFORM_E2E_CHECKLIST.md).
