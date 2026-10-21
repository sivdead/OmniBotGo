# Real-platform E2E checklist

Use when live credentials are available. Do **not** commit secrets.

## Preconditions

- App running with `HTTP_API_KEY` set; dialect migrations applied
- Admin API: `Authorization: Bearer $HTTP_API_KEY`
- Platform credentials provided out-of-band (env or secure store)

## Cases

1. **Multi-channel stream isolation (DingTalk Stream / Telegram)**  
   Two channels, distinct credentials → both connected; inbound `channel_id` correct; AI history does not cross channels for the same platform conversation id.

2. **Webhook handshake + inbound**  
   WeCom / WeChat Official / Feishu URL verification; POST inbound text → stored pending→processed; ACK only after persist/process rules.

3. **Unsupported outbound type**  
   Send a type not in the platform’s supported set → API/adapter error; message not marked sent as text.

4. **Create-time config validation**  
   Create channel with missing required fields → rejected; valid config accepted; webhook path `/webhook/{platform}/{uuid}`.

5. **Troubleshoot + retry**  
   Force a processor failure → `GET /api/v1/messages/{id}/troubleshoot` shows `stage=process`, `recoverable=true`; `POST …/retry` recovers without unique-key collision. Repeat for outbound deliver failure (`stage=deliver`).

6. **Reply-to-original**  
   DM and group inbound → `POST /messages/{id}/reply` addresses the correct conversation.
