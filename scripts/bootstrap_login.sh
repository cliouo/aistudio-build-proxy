#!/usr/bin/env bash
set -e
PORT=${CFX_PORT:-38835}
python -m camoufox server --port "$PORT" &
server_pid=$!
trap 'kill $server_pid' EXIT
export CAMOUFOX_WS="ws://localhost:$PORT"

echo "👉 请在弹出的 Firefox 中完成 Google 登录和 2FA，然后按回车继续 …"
go run ./cmd/agent --bootstrap --ws-endpoint "$CAMOUFOX_WS" --url "$BROWSER_URL"
