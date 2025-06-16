#!/usr/bin/env bash
set -e
export PYTHONUNBUFFERED=1

PORT=${CFX_PORT:-38835}
LOG=/tmp/camoufox.log
rm -f "$LOG"

cmd=(python -m camoufox server --virtual-display)
if [[ -n "$USE_XVFB" ]]; then
  cmd=(xvfb-run --server-arg="-screen 0 1920x1080x24" "${cmd[@]}")
fi
if [[ "$PORT" != "0" ]]; then
  cmd+=(--port "$PORT")
fi

"${cmd[@]}" | tee "$LOG" &
pid=$!

if [[ "$PORT" = "0" ]]; then
  while ! grep -q "Websocket endpoint:" "$LOG"; do sleep 0.2; done
  grep "Websocket endpoint:" "$LOG" | awk '{print $3}' > /tmp/endpoint.txt
else
  echo "ws://localhost:${PORT}" > /tmp/endpoint.txt
fi

wait $pid
