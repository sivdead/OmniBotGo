#!/usr/bin/env bash
# Smoke: A2A client against a live agent (default http://127.0.0.1:10000).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
BASE="${A2A_BASE_URL:-http://127.0.0.1:10000}"

echo "== Agent Card =="
curl -fsS --max-time 5 "${BASE}/.well-known/agent.json" | head -c 500
echo
echo

echo "== Go client SendText =="
go run ./scripts/smoke_a2a_main.go -base "$BASE" -text "${1:-How much is 10 USD to EUR?}"
