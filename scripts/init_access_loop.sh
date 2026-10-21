#!/usr/bin/env bash
# Copyable access-loop smoke (T09): bot → channel → processor → fallback rule → routing preview.
# Does not call real platforms. Requires a running OmniBotGo with HTTP_API_KEY.
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
HTTP_API_KEY="${HTTP_API_KEY:?HTTP_API_KEY is required}"
AUTH="Authorization: Bearer ${HTTP_API_KEY}"
SUFFIX="$(date +%s)-$$"

echo "== wait /ready =="
for i in $(seq 1 30); do
  code="$(curl -sS -o /dev/null -w '%{http_code}' "$BASE_URL/ready" || true)"
  if [[ "$code" == "200" ]]; then
    break
  fi
  sleep 1
done
[[ "${code:-}" == "200" ]] || { echo "ready check failed (last=$code)"; exit 1; }

echo "== create bot =="
BOT_JSON="$(curl -sS -X POST "$BASE_URL/api/v1/bots" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"bot_name\":\"access-loop-$SUFFIX\",\"bot_type\":\"general\",\"description\":\"T09\",\"created_by\":\"script\"}")"
BOT_ID="$(printf '%s' "$BOT_JSON" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
[[ -n "$BOT_ID" ]] || { echo "bot create failed: $BOT_JSON"; exit 1; }
echo "BOT_ID=$BOT_ID"

echo "== create channel =="
CH_JSON="$(curl -sS -X POST "$BASE_URL/api/v1/channels" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"bot_id\":\"$BOT_ID\",\"platform_type\":\"wecom\",\"channel_name\":\"ch-$SUFFIX\",\"config\":{\"corp_id\":\"demo\",\"agent_id\":\"1\",\"secret\":\"demo-secret\"}}")"
CHANNEL_ID="$(printf '%s' "$CH_JSON" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
[[ -n "$CHANNEL_ID" ]] || { echo "channel create failed: $CH_JSON"; exit 1; }
echo "CHANNEL_ID=$CHANNEL_ID"

echo "== create processor =="
PROC_JSON="$(curl -sS -X POST "$BASE_URL/api/v1/processors" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"name\":\"proc-$SUFFIX\",\"type\":\"auto_reply\",\"config\":{\"reply_content\":\"pong\",\"api_key\":\"should-redact\"},\"priority\":10}")"
PROCESSOR_ID="$(printf '%s' "$PROC_JSON" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
[[ -n "$PROCESSOR_ID" ]] || { echo "processor create failed: $PROC_JSON"; exit 1; }
echo "PROCESSOR_ID=$PROCESSOR_ID"
if printf '%s' "$PROC_JSON" | grep -q 'should-redact'; then
  echo "WARN: processor response echoed api_key (expected redacted)"
fi

echo "== create fallback rule =="
RULE_JSON="$(curl -sS -X POST "$BASE_URL/api/v1/processors/$PROCESSOR_ID/rules" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"rule_name\":\"fallback-$SUFFIX\",\"route_type\":1,\"priority\":100,\"is_fallback\":true,\"status\":1}")"
echo "$RULE_JSON" | head -c 200; echo

echo "== routing preview =="
PREV="$(curl -sS -X POST "$BASE_URL/api/v1/routing/preview" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"channel_id\":\"$CHANNEL_ID\",\"message_type\":\"text\",\"content\":\"hi\",\"sender_id\":\"u1\"}")"
echo "$PREV"

echo "OK access loop scaffolded. Configure platform webhook to channel webhook_path; use /messages/send or /messages/{id}/reply."
